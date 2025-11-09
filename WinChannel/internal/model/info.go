package model

type InfoResponse struct {
    HostIP string   `json:"host_ip"`
    Port   int      `json:"port"`
    Urls   []string `json:"urls"`
}