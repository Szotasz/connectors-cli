package api

type Manifest struct {
	Connectors []ConnectorInfo `json:"connectors"`
	Tools      []ToolEntry     `json:"tools"`
}

type ConnectorInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ToolEntry struct {
	Connector     string `json:"connector"`
	ConnectorName string `json:"connector_name"`
	Command       string `json:"command"`
	Description   string `json:"description"`
	Args          []Arg  `json:"args"`
}

type Arg struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}
