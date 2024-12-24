package c64dws

// Server events
type ServerEventType string

const (
	ServerEventTypeBreakpoint ServerEventType = "breakpoint"
)

// Breakpoint events
type BreakpointEventType string

const (
	BreakpointEventRaster  BreakpointEventType = "rasterLine"
	BreakpointEventCPUData BreakpointEventType = "data"
	BreakpointEventCPUAddr BreakpointEventType = "addr"
)

// Memory breakpoint access types
type MemoryBreakpointAccess string

const MemoryBreakpointAccessRead MemoryBreakpointAccess = "read"
const MemoryBreakpointAccessWrite MemoryBreakpointAccess = "write"

// Server event base
type ServerEventBase struct {
	Event string `json:"event"`
}

// Breakpoint event base
type BreakpointEventBase struct {
	ServerEventBase
	Type         string `json:"type"`
	BreakpointId uint64 `json:"breakpointId"`
	Platform     string `json:"platform"`
}

// Raster breakpoint event
type RasterBreakpointEvent struct {
	BreakpointEventBase
	RasterLine uint16 `json:"rasterLine"`
}

// CPU Address breakpoint event
type CPUAddrBreakpointEvent struct {
	BreakpointEventBase
	Segment uint16 `json:"segment"`
}

// CPU Data breakpoint event
type CPUDataBreakpointEvent struct {
	BreakpointEventBase
	Segment uint16 `json:"segment"`
}
