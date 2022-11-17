package k8sObjManager

type IPs struct {
	InterfaceName string
	IPv4          string
	IPv6          string
}
type PodIps map[string][]IPs

type MultusAnnotationValueItem struct {
	Interface string   `json:"interface"`
	Ips       []string `json:"ips"`
}
