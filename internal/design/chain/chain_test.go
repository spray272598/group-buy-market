package chain

import (
	"context"
	"testing"
)

type ctxData struct {
	steps []string
}

type nodeA struct {
	BaseHandler[string, ctxData, string]
}

func (n *nodeA) Apply(ctx context.Context, req string, dc *ctxData) (string, error) {
	dc.steps = append(dc.steps, "A")
	return n.Next(ctx, req, dc)
}

type nodeB struct {
	BaseHandler[string, ctxData, string]
}

func (n *nodeB) Apply(ctx context.Context, req string, dc *ctxData) (string, error) {
	dc.steps = append(dc.steps, "B")
	return req + "-ok", nil
}

func TestLinkedList(t *testing.T) {
	list := NewLinkedList[string, ctxData, string]("test", &nodeA{}, &nodeB{})
	dc := &ctxData{}
	res, err := list.Apply(context.Background(), "hello", dc)
	if err != nil {
		t.Fatal(err)
	}
	if res != "hello-ok" {
		t.Fatalf("got %s", res)
	}
	if len(dc.steps) != 2 || dc.steps[0] != "A" || dc.steps[1] != "B" {
		t.Fatalf("steps %v", dc.steps)
	}
}
