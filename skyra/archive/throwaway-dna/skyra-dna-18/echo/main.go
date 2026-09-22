package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	p "skyra-v1/src/skyra/skyra-dna-18/protocol"
)

func serve(in io.Reader, out io.Writer) error {
	decoder, encoder := json.NewDecoder(in), json.NewEncoder(out)
	for {
		var m p.Move
		if err := decoder.Decode(&m); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := m.Validate(); err != nil {
			return err
		}
		if m.Target != p.Echo || m.Invoke == nil {
			return fmt.Errorf("echo expects Invoke")
		}
		if err := p.CheckArguments(m.Invoke.Args, "echo"); err != nil {
			return err
		}
		result := p.Move{Identity: m.Identity, Context: m.Context, Source: m.Invoke.Frame, Return: &p.Return{Answering: m.Invoke.Frame, Observation: m.Invoke.Args[1]}}
		if err := encoder.Encode(result); err != nil {
			return err
		}
	}
}

func main() {
	if err := serve(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
