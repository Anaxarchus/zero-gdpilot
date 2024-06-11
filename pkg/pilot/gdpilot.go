package pilot

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// StandardChans holds channels for communication between the GodotPilot and the Go application.
type StandardChans struct {
	In   chan string
	Out  chan string
	Err  chan string
	Quit chan struct{}
}

// NewStandardChans initializes and returns a new StandardChans instance.
func NewStandardChans() *StandardChans {
	return &StandardChans{
		In:   make(chan string),
		Out:  make(chan string),
		Err:  make(chan string),
		Quit: make(chan struct{}),
	}
}

// StandardPipes holds the standard input, output, and error pipes.
type StandardPipes struct {
	In  io.WriteCloser
	Out io.ReadCloser
	Err io.ReadCloser
}

// GodotPilot manages the execution of a Godot executable and handles its I/O streams.
type GodotPilot struct {
	executablePath string
	Cmd            *exec.Cmd
	Chans          *StandardChans
	Pipes          *StandardPipes
	OnStdOut       func(line string)
	OnStdErr       func(line string)
}

// NewGodotPilot initializes and returns a new GodotPilot instance.
func NewGodotPilot(onStdOut func(line string), onStdErr func(line string), executablePath string, arguments ...string) *GodotPilot {
	pilot := &GodotPilot{
		executablePath: executablePath,
		Cmd:            exec.Command(executablePath, arguments...),
		Chans:          NewStandardChans(),
		OnStdOut:       onStdOut,
		OnStdErr:       onStdErr,
	}
	pipes := &StandardPipes{}
	var err error
	pipes.In, err = pilot.Cmd.StdinPipe()
	if err != nil {
		// handle error appropriately
		fmt.Printf("Error creating stdin pipe: %v\n", err)
		return nil
	}
	pipes.Out, err = pilot.Cmd.StdoutPipe()
	if err != nil {
		// handle error appropriately
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		return nil
	}
	pipes.Err, err = pilot.Cmd.StderrPipe()
	if err != nil {
		// handle error appropriately
		fmt.Printf("Error creating stderr pipe: %v\n", err)
		return nil
	}
	pilot.Pipes = pipes
	return pilot
}

// Start begins the execution of the Godot executable and starts goroutines to handle its I/O streams.
func (gp *GodotPilot) Start() error {
	err := gp.Cmd.Start()
	if err != nil {
		return err
	}
	go gp.handleStdIn()
	go gp.handleStdOut()
	go gp.handleStdErr()
	return nil
}

// handleStdIn manages writing commands to the standard input of the Godot executable.
func (gp *GodotPilot) handleStdIn() {
	defer gp.Pipes.In.Close()
	for {
		select {
		case command, ok := <-gp.Chans.In:
			if !ok {
				return
			}
			_, err := gp.Pipes.In.Write([]byte(command + "\n"))
			if err != nil {
				fmt.Printf("Error writing to stdin: %v\n", err)
				return
			}
		case <-gp.Chans.Quit:
			return
		}
	}
}

// handleStdOut reads from the standard output of the Godot executable and processes each line.
func (gp *GodotPilot) handleStdOut() {
	defer close(gp.Chans.Out)
	reader := bufio.NewReader(gp.Pipes.Out)
	for {
		select {
		case <-gp.Chans.Quit:
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading stdout: %v\n", err)
				}
				return
			}
			if len(line) == 0 {
				continue
			}
			gp.OnStdOut(strings.TrimSpace(line))
		}
	}
}

// handleStdErr reads from the standard error of the Godot executable and processes each line.
func (gp *GodotPilot) handleStdErr() {
	defer close(gp.Chans.Err)
	reader := bufio.NewReader(gp.Pipes.Err)
	for {
		select {
		case <-gp.Chans.Quit:
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading stderr: %v\n", err)
				}
				return
			}
			if len(line) == 0 {
				continue
			}
			gp.OnStdErr(strings.TrimSpace(line))
		}
	}
}
