package internal

const (
	ACTION_SET_WARN = "set-warn"
	ACTION_SET_ERR  = "set-err"
	ACTION_BEAT     = "beat"
)

type Cmd struct {
	Action  string
	Service string
	Value   int
}

type ViewCmd struct {
	Action string
	View   string
}
