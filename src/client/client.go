package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/jroimartin/gocui"
)

var (
	connection net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
)

func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if messages, err := g.SetView("messages", 1, 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		messages.Title = " messages: "
		messages.Autoscroll = true
		messages.Wrap = true
	}

	if input, err := g.SetView("input", 1, maxY-5, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		input.Title = " send: "
		input.Autoscroll = false
		input.Wrap = true
		input.Editable = true
	}

	if name, err := g.SetView("name", maxX/2-10, maxY/2-1, maxX/2+10, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetCurrentView("name")
		name.Title = "  name: "
		name.Autoscroll = false
		name.Wrap = true
		name.Editable = true
	}
	return nil
}

func Disconnect(g *gocui.Gui, v *gocui.View) error {
	connection.Close()
	return gocui.ErrQuit
}

func Send(g *gocui.Gui, v *gocui.View) error {
	writer.WriteString(v.Buffer())
	writer.Flush()
	g.Execute(func(g *gocui.Gui) error {
		v.Clear()
		v.SetCursor(0, 0)
		v.SetOrigin(0, 0)
		return nil
	})
	return nil
}

// Connect to the server, create new reader, writer
// set client name
func Connect(g *gocui.Gui, v *gocui.View) error {
	conn, err := net.Dial("tcp", ":5000")
	if err != nil {
		log.Fatal(err)
	}
	connection = conn
	reader = bufio.NewReader(connection)
	writer = bufio.NewWriter(connection)
	writer.WriteString("/name>" + v.Buffer())
	writer.Flush()
	// Some UI changes
	g.SetViewOnTop("messages")
	g.SetViewOnTop("input")
	g.SetCurrentView("input")
	// Wait for server messages in new goroutine
	messagesView, _ := g.View("messages")
	go func() {
		for {
			data, _ := reader.ReadString('\n')
			msg := strings.TrimSpace(data)
			g.Execute(func(g *gocui.Gui) error {
				fmt.Fprintln(messagesView, msg)
				return nil
			})
		}
	}()
	return nil
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()

	g.SetManagerFunc(Layout)

	g.SetKeybinding("name", gocui.KeyEnter, gocui.ModNone, Connect)

	g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, Send)

	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, Disconnect)

	g.MainLoop()
}