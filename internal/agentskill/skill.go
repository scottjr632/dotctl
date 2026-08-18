package agentskill

import _ "embed"

// Content is the agent skill installed by dotctl agent install-skill.
//
//go:embed dotctl/SKILL.md
var Content string
