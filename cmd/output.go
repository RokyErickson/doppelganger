package cmd

import (
	"fmt"

	"github.com/fatih/color"
)

type StatusLinePrinter struct {
	nonEmpty bool
}

func (p *StatusLinePrinter) Print(message string) {

	fmt.Fprintf(color.Output, statusLineFormat, message)

	p.nonEmpty = true
}

func (p *StatusLinePrinter) Clear() {

	p.Print("")

	fmt.Print("\r")

	p.nonEmpty = false
}

func (p *StatusLinePrinter) BreakIfNonEmpty() {

	if p.nonEmpty {
		fmt.Println()
		p.nonEmpty = false
	}
}
