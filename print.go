package jiji

type Print struct{}

func (t Print) Connect() error {
	return nil
}

func (t Print) Close() {
	return
}

func (t Print) Send(data []byte) error {
	Logger.Println("Send:", string(data))
	return nil
}
