package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/laktak/lterm"
	"golang.org/x/term"
)

const (
	escMoveRight = "\033[C"
)

var reChk = regexp.MustCompile(`(..)-.+=(\d+)/(\d+)`)

type Rstyle struct {
	bg      string
	dim     string
	text    string
	bar1    string
	bar2    string
	spin    string
	spinner []string
}

type Rsyncy struct {
	style       Rstyle
	trans       string
	percent     float64
	speed       string
	xfr         string
	chk         string
	chkFinished bool
	start       time.Time
	statusOnly  bool
}

func NewRsyncy(rstyle Rstyle) *Rsyncy {
	return &Rsyncy{
		style: rstyle,
		start: time.Now(),
	}
}

func (r *Rsyncy) parseRsyncStat(line string) bool {
	// sample: 3.93M   5%  128.19kB/s    0:00:29 (xfr#208, ir-chk=2587/2821)
	// sample: 130.95M  29%  207.03kB/s    0:10:17 (xfr#4000, to-chk=1000/5055)
	data := strings.Fields(line)

	if len(data) >= 4 &&
		strings.HasSuffix(data[1], "%") {
		r.trans = data[0]
		if p, err := strconv.ParseFloat(strings.TrimSuffix(data[1], "%"), 64); err == nil {
			r.percent = p / 100.0
		} else {
			// skip
			log.Printf("ERROR - can't parse#1: '%s' in %s\n", data[1], line)
		}
		r.speed = data[2]
		// ignore data[3] (time)

		if len(data) == 6 {
			xfr := strings.Split(strings.TrimSuffix(data[4], ","), "#")
			if len(xfr) == 2 {
				r.xfr = "#" + xfr[1]
			}

			match := reChk.FindStringSubmatch(data[5])
			if len(match) == 4 {
				r.chkFinished = match[1] == "to"
				todo, errTodo := strconv.Atoi(match[2])
				total, errTotal := strconv.Atoi(match[3])
				if errTodo == nil && errTotal == nil && total > 0 {
					done := total - todo
					chkPercent := float64(done) * 100.0 / float64(total)
					r.chk = fmt.Sprintf("%2.0f%% (%d)", chkPercent, total)
				} else {
					r.chk = ""
				}
			} else {
				r.chk = ""
			}
		}
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func (r *Rsyncy) drawStat() {
	cols := lterm.GetWidth()
	elapsed := time.Since(r.start)
	elapsedStr := formatDuration(elapsed)

	spin := ""
	if !r.chkFinished && len(r.style.spinner) > 0 {
		spin = r.style.spinner[int(elapsed.Seconds())%len(r.style.spinner)]
	}

	// define status (excl. bar)
	// use \xff as a placeholder for the spinner
	parts := []string{
		fmt.Sprintf("%11s", r.trans),
		fmt.Sprintf("%14s", r.speed),
		elapsedStr,
		r.xfr,
		fmt.Sprintf("scan %s\xff", r.chk),
	}

	// reduce to fit
	plen := func(parts []string) int {
		sum := 0
		for _, s := range parts {
			sum += len(s)
		}
		return sum + len(parts)
	}

	for len(parts) > 0 && plen(parts) > cols {
		parts = parts[1:]
	}

	// add bar in remaining space
	pc := 0
	rcols := cols - plen(parts)
	pcStr := ""
	if rcols > 12 {
		pcWidth := min(rcols-7, 30)
		pc = int(r.percent * float64(pcWidth))
		pcStr = fmt.Sprintf("%s[%s%s%s%s]%s%4.0f%%",
			r.style.bar1, r.style.bar2, strings.Repeat("#", pc),
			r.style.bar1, strings.Repeat(":", pcWidth-pc),
			r.style.text, r.percent*100,
		)
		rcols -= pcWidth + 7
	} else if rcols > 5 {
		pcStr = fmt.Sprintf("%5.0f%%", r.percent*100)
		rcols -= 5
	}
	if pcStr != "" {
		parts = append([]string{pcStr}, parts...)
	}

	// get delimiter size
	delim := fmt.Sprintf("%s|%s", r.style.dim, r.style.text)
	if rcols > (len(parts)-1)*2 {
		delim = " " + delim + " "
	}

	// render with delimiter
	status := strings.Replace(strings.Join(parts, delim), "\xff", r.style.spin+spin+r.style.text, 1)

	if !r.statusOnly {
		// write and position cursor on bar
		lterm.Write("\r", r.style.bg, status, lterm.ClearLine(0), "\r", strings.Repeat(escMoveRight, pc), lterm.Reset)
	} else {
		lterm.Write("\r\n", r.style.bg, status, lterm.Reset)
	}
}

func (r *Rsyncy) parseLine(lineBytes []byte, isStatHint bool) {
	line := string(bytes.TrimSpace(lineBytes))
	if line == "" {
		return
	}

	isStat := isStatHint || bytes.HasPrefix(lineBytes, []byte("\r"))
	line = strings.TrimPrefix(line, "\r")

	if isStat {
		if !r.parseRsyncStat(line) && !r.statusOnly {
			lterm.Printline("\r", line)
		}
		r.drawStat()
	} else if strings.HasSuffix(line, "/") {
		// skip directories
	} else if !r.statusOnly {
		lterm.Printline("\r", line)
		r.drawStat()
	}
}

func (r *Rsyncy) readOutput(reader io.Reader) {
	bufReader := bufio.NewReader(reader)
	var lineBuffer bytes.Buffer
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	inputChan := make(chan byte, 1024)
	errChan := make(chan error, 1)

	// read bytes and send them to the channel
	go func() {
		for {
			b, err := bufReader.ReadByte()
			if err != nil {
				errChan <- err // signal error (including io.EOF)
				return
			}
			inputChan <- b
		}
	}()

	for {
		select {
		case b := <-inputChan:
			ticker.Reset(200 * time.Millisecond)

			if b == '\r' {
				r.parseLine(lineBuffer.Bytes(), true)
				lineBuffer.Reset()
				lineBuffer.WriteByte('\r')
			} else if b == '\n' {
				r.parseLine(lineBuffer.Bytes(), false)
				lineBuffer.Reset()
			} else {
				lineBuffer.WriteByte(b)
			}

		case <-ticker.C:
			// no new input
			if lineBuffer.Len() > 0 {
				// assume this is a status update
				r.parseLine(lineBuffer.Bytes(), true)
				lineBuffer.Reset()
			} else {
				r.drawStat()
			}

		case <-errChan:
			// exit
			r.parseLine(lineBuffer.Bytes(), false)
			lterm.Printline("\r")
			return
		}
	}
}

func runRsync(args []string, wg *sync.WaitGroup, rcChan chan<- int, stdoutWriter io.WriteCloser) {
	defer wg.Done()
	defer stdoutWriter.Close()

	// prefix rsync and add args required for progress
	args = append(args, "--info=progress2", "--no-v", "-hv")

	cmd := exec.Command("rsync", args...)
	cmd.Stdout = stdoutWriter
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			log.Printf("Error running rsync: %v", err)
			exitCode = 1
		}
	}
	rcChan <- exitCode
}

func main() {
	log.SetFlags(0)

	var rstyle Rstyle
	var spinner = []string{"-", "\\", "|", "/"}
	if lterm.GetTermColorBits() >= 8 {
		rstyle = Rstyle{
			bg:      lterm.Bg8(238),
			dim:     lterm.Fg8(241),
			text:    lterm.Fg8(250),
			bar1:    lterm.Fg8(243),
			bar2:    lterm.Fg8(43),
			spin:    lterm.Fg8(228) + lterm.Bold,
			spinner: spinner,
		}
	} else {
		rstyle = Rstyle{
			bg:      lterm.Bg4(7),
			dim:     lterm.Fg4(8),
			text:    lterm.Fg4(0),
			bar1:    lterm.Fg4(8),
			bar2:    lterm.Fg4(0),
			spin:    lterm.Fg4(14) + lterm.Bold,
			spinner: spinner,
		}
	}

	rsyncy := NewRsyncy(rstyle)

	// status only mode (use with lf)
	if strings.HasSuffix(os.Args[0], "rsyncy-stat") {
		rsyncy.statusOnly = true
	}

	// handle ctrl+c
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		lterm.Write(lterm.Reset, "\r", lterm.ClearLine(0), "\r\naborted\r\n")
		os.Exit(1)
	}()

	args := os.Args[1:]

	if len(args) == 0 {
		stdinFd := int(os.Stdin.Fd())
		if term.IsTerminal(stdinFd) {
			fmt.Println("rsyncy is an rsync wrapper with a progress bar.")
			fmt.Println("Please specify your rsync options as you normally would but use rsyncy instead of rsync.")
		} else {
			// receive pipe from rsync
			rsyncy.readOutput(os.Stdin)
		}
	} else {
		readPipe, writePipe, err := os.Pipe()
		if err != nil {
			log.Fatalf("error creating pipe: %v", err)
			os.Exit(1)
		}

		rcChan := make(chan int, 1)
		var wg sync.WaitGroup
		wg.Add(1)

		go runRsync(args, &wg, rcChan, writePipe)
		rsyncy.readOutput(readPipe)
		wg.Wait()
		exitCode := <-rcChan
		os.Exit(exitCode)
	}
}
