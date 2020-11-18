#!/usr/bin/env python3

import os
import queue
import re
import select
import subprocess
import sys
import time
import types
from threading import Thread
from datetime import datetime

_re_chk = re.compile(r"(..)-.+=(\d+)/(\d+)")

# start inline from laktakpy


class CLI:
    class style:
        reset = "\033[0m"
        bold = "\033[01m"

    class esc:
        right = "\033[C"

        def clear_line(opt=0):
            # 0=to end, 1=from start, 2=all
            return "\033[" + str(opt) + "K"

    def get_col_bits():
        c, t = os.environ.get("COLORTERM", ""), os.environ.get("TERM", "")
        if c in ["truecolor", "24bit"]:
            return 24
        elif c == "8bit" or "256" in t:
            return 8
        else:
            return 4

    # 4bit system colors
    def fg4(col):
        # black=0,red=1,green=2,orange=3,blue=4,purple=5,cyan=6,lightgrey=7
        # darkgrey=8,lightred=9,lightgreen=10,yellow=11,lightblue=12,pink=13,lightcyan=14
        return f"\033[{(30+col) if col<8 else (90-8+col)}m"

    def bg4(col):
        # black=0,red=1,green=2,orange=3,blue=4,purple=5,cyan=6,lightgrey=7
        return f"\033[{40+col}m"

    # 8bit xterm colors
    def fg8(col):
        return f"\033[38;5;{col}m"

    def bg8(col):
        return f"\033[48;5;{col}m"

    def write(*text):
        for t in text:
            sys.stdout.write(str(t))
        sys.stdout.flush()

    def printline(*text, sep=" ", end="\r\n"):
        CLI.write("\r", sep.join([str(t) for t in text]), CLI.esc.clear_line(), end)

    def get_size():
        # returns {columns, lines}
        if sys.stdout.isatty():
            return os.get_terminal_size()
        else:
            return types.SimpleNamespace(columns=80, lines=40)


# end inline from laktakpy


class Rsyncy:
    def __init__(self, rstyle):
        self.bg = rstyle["bg"]
        self.cdim = rstyle["dim"]
        self.ctext = rstyle["text"]
        self.cbar1 = rstyle["bar1"]
        self.cbar2 = rstyle["bar2"]
        self.cspin = rstyle["spin"]
        self.spinner = rstyle["spinner"]
        self.trans = 0
        self.percent = 0
        self.speed = ""
        self.xfr = ""
        self.chk = ""
        self.chk_finished = False
        self.start = datetime.now()

    def parse_stat(self, line):
        # sample: 6,672,528  96%    1.04MB/s    0:00:06 (xfr#1, to-chk=7/12)
        data = [s for s in line.split(" ") if s]
        if len(data) >= 4:
            self.trans, percent, self.speed, timing, *_ = data
            try:
                self.percent = int(percent.strip("%")) / 100
            except Exception as e:
                print("ERROR - can't parse#1:", line, data, e, sep="\n> ")

        # timing is remaining with 4 args, or elapsed with 6
        if len(data) == 6:
            try:
                xfr, chk = data[4:6]
                self.xfr = xfr.strip(",").split("#")[1]
                if self.xfr:
                    self.xfr = "#" + self.xfr

                m = _re_chk.match(chk)
                if m:
                    self.chk_finished = m[1] == "to"
                    todo = int(m[2])
                    total = int(m[3])
                    done = total - todo
                    self.chk = f"{(done/total if total else 0):2.0%} ({total})"
                else:
                    self.chk = ""

            except Exception as e:
                print("ERROR - can't parse#2:", line, data, e, sep="\n> ")

    def draw_stat(self):
        cols = CLI.get_size().columns
        elapsed = datetime.now() - self.start
        if self.chk_finished:
            spin = ""
        else:
            spin = self.spinner[round(elapsed.total_seconds()) % len(self.spinner)]

        # define status (excl. bar)
        # use \xff as a placeholder for the spinner
        parts = [
            o
            for o in [
                f"{self.trans:>11}",
                f"{self.speed:>14}",
                f"{str(elapsed).split('.')[0]}",
                f"{self.xfr}",
                f"scan {self.chk}\xff",
            ]
            if o
        ]

        # reduce to fit
        plen = lambda: sum(len(s) for s in parts) + len(parts)
        while parts and plen() > cols:
            parts.pop(0)

        # add bar in remaining space
        pc = 0
        rcols = cols - plen()
        if rcols > 12:
            pc_width = min(rcols - 7, 30)
            pc = round(self.percent * pc_width)
            parts.insert(
                0,
                f"{self.cbar1}[{self.cbar2}{'#' * pc}{self.cbar1}{':' * (pc_width-pc)}]{self.ctext}"
                + f"{self.percent:>5.0%}",
            )
            rcols -= pc_width + 7
        elif rcols > 5:
            parts.insert(0, f"{self.percent:>5.0%}")
            rcols -= 5

        # get delimiter size
        delim = f"{self.cdim}|{self.ctext}"
        if rcols > (len(parts) - 1) * 2:
            delim = " " + delim + " "

        # render with delimiter
        status = delim.join(parts).replace("\xff", f"{self.cspin}{spin}{self.ctext}")

        # write and position cursor on bar
        CLI.write(
            "\r",
            self.bg,
            status,
            CLI.esc.clear_line(),
            "\r",
            f"{CLI.esc.right * pc}",
            CLI.style.reset,
        )

    def parse_line(self, line, is_stat):
        line = line.decode().strip(" ")
        if not line:
            return

        is_stat = is_stat or line[0] == "\r"
        line = line.replace("\r", "")

        if is_stat:
            self.parse_stat(line)
            self.draw_stat()
        elif line[-1] == "/":
            # skip directories
            pass
        else:
            CLI.printline(line)
            self.draw_stat()

    def read(self, fd):
        line = b""
        while True:
            stat = select.select([fd], [], [], 0.2)
            if fd in stat[0]:
                ch = os.read(fd, 1)
                if ch == b"":
                    # exit
                    self.parse_line(line, False)
                    break

                elif ch == b"\r":
                    self.parse_line(line, True)
                    line = b"\r"
                elif ch == b"\n":
                    self.parse_line(line, False)
                    line = b""
                else:
                    line += ch

            else:
                # no new input
                if line:
                    # assume this is a status update
                    self.parse_line(line, True)
                    line = b""
                else:
                    # waiting for input
                    self.draw_stat()
                    time.sleep(0.5)

        CLI.printline("")


def run_rsync(args, write_pipe, rc):
    # prefix rsync and add args required for progress
    args = ["rsync"] + args + ["--info=progress2", "--no-v", "-hv"]
    try:
        p = subprocess.Popen(args, stdout=write_pipe)
        p.wait()
    finally:
        os.close(write_pipe)
        rc.put(p.returncode)


if __name__ == "__main__":

    if CLI.get_col_bits() >= 8:
        rstyle = {
            "bg": CLI.bg8(238),
            "dim": CLI.fg8(241),
            "text": CLI.fg8(250),
            "bar1": CLI.fg8(243),
            "bar2": CLI.fg8(43),
            "spin": CLI.fg8(228) + CLI.style.bold,
            "spinner": ["-", "\\", "|", "/"],
        }
    else:
        rstyle = {
            "bg": CLI.bg4(7),
            "dim": CLI.fg4(8),
            "text": CLI.fg4(0),
            "bar1": CLI.fg4(8),
            "bar2": CLI.fg4(0),
            "spin": CLI.fg4(14) + CLI.style.bold,
            "spinner": ["-", "\\", "|", "/"],
        }
    rsyncy = Rsyncy(rstyle)

    try:

        if len(sys.argv) == 1:
            if sys.stdin.isatty():
                print("rsyncy is an rsync wrapper with a progress bar.")
                print(
                    "Please specify your rsync options as you normally would but use rsyncy instead of rsync."
                )
            else:
                # receive pipe from rsync
                rsyncy.read(sys.stdin.fileno())
        else:
            read_pipe, write_pipe = os.pipe()
            rc = queue.Queue()
            t = Thread(target=run_rsync, args=(sys.argv[1:], write_pipe, rc))
            t.start()
            rsyncy.read(read_pipe)
            t.join()
            sys.exit(rc.get())

    except KeyboardInterrupt:
        pass
