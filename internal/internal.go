package internal

const (
	ACTION_SET_WARN = "set-warn"
	ACTION_SET_ERR  = "set-err"
	ACTION_BEAT     = "beat"
)

const (
	ACTION_REFRESH_VIEW = "refresh-view"
	ACTION_UPSERT_VIEW  = "upsert-view"
)

type Cmd struct {
	Action  string
	Service string
	Value   int
}

type ViewCmd struct {
	Action string
	View   string
	Regexp string
}
