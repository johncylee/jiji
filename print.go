package jiji

import "context"

// Print everything to Logger.Info
type Print struct{}

func (t Print) Connect(ctx context.Context) error {
	return nil
}

func (t Print) Close() {
}

func (t Print) Send(data []byte, ctx context.Context) error {
	Logger.Info("Send", "msg", string(data))
	return nil
}
