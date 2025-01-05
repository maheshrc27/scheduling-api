package transfer

type PostCreation struct {
	Caption          string `json:"caption"`
	Title            string `json:"title"`
	ScheduledTime    string `json:"scheduled_time"`
	SelectedAccounts string `json:"selected_account"`
}
