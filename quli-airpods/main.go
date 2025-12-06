package main

type Airpod struct {
	//TODO
}

type AirpodCase struct {
	//TODO

}

func NewAirpodCase() *AirpodCase {
	return nil
}

func (a *AirpodCase) GetRightAirpod() *Airpod {
	return nil
}

func (a *AirpodCase) GetLeftAirpod() *Airpod {
	return nil

}
func (a *Airpod) GetState() string {
	return ""
}

func (a *AirpodCase) UndockLeft() *Airpod {
	return nil
}

func (a *AirpodCase) UndockRight() *Airpod {
	return nil
}

func (a *AirpodCase) DockLeft() error {
	return nil

}
func (a *AirpodCase) DockRight() error {
	return nil
}

func (a *Airpod) GetChannel() chan byte {
	return nil
}

func (c *AirpodCase) ConnectBluetooth(ch chan byte) error {
	return nil
}
