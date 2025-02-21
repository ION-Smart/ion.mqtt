package cvevents

type MessageCrowd struct {
	Count           int     `json:"count"`
	FrameId         int     `json:"frame_id"`
	FrameTime       float32 `json:"frame_time"`
	InstanceId      string  `json:"instance_id"`
	SystemDate      string  `json:"system_date"`
	SystemTimestamp string  `json:"system_timestamp"`
}

type MessageSecuRT struct {
	Events []struct {
		Id         string `json:"id"`
		ZoneId     string `json:"zone_id"`
		InstanceId string `json:"instance_id"`
		Type       string `json:"type"`
		Extra      struct {
			CrowdingMinCount int `json:"crowding_min_count"`
			CurrentEntries   int `json:"current_entries"`
			TotalHits        int `json:"total_hits"`
		} `json:"extra"`
	} `json:"events"`
	FrameId         int     `json:"frame_id"`
	FrameTime       float32 `json:"frame_time"`
	SystemDate      string  `json:"system_date"`
	SystemTimestamp string  `json:"system_timestamp"`
}
