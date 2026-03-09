package jiji

type Print struct{}

func (t Print) Connect() error {
	return nil
}

func (t Print) Close() {
}

func (t Print) Send(data []byte) error {
	Logger.Info("Send", "msg", string(data))
	return nil
}
