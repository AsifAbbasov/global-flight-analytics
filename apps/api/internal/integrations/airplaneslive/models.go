package airplaneslive

type StateResponse struct {
	Time     int64   `json:"time"`
	States   [][]any `json:"ac"`
	Messages int     `json:"msg"`
	Total    int     `json:"total"`
}
