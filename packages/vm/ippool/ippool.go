package ippool

type IPStatus string

const (
	IPFree     IPStatus = "ideal"
	IPLeased   IPStatus = "leased"
	IPDraining IPStatus = "draining"
)

type pool struct {
	
}
