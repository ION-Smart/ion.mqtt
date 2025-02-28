package models

import nx "github.com/ION-Smart/ion.mqtt/pkg/nxwitness"

type DispositivoCloud struct {
	CodDispositivo int
	NomDispositivo string
	DeviceId       string
	Cloud          *nx.NxCloud
	SystemId       string
	CloudBaseUser  string
	CloudBasePass  string
	Ip             string
	Puerto         int
	Server         string
}

func (disp *DispositivoCloud) ObtenerImagen(timestamp int64) (string, error) {
	return nx.GetDeviceThumbnailB64(disp.Cloud, "jpg", disp.DeviceId, timestamp)
}
