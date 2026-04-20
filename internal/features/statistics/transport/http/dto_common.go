package statistics_transport_http

import (
	"github.com/LisLisich/RESTAPI/internal/core/domain"
)

type StatisticsDTOResponse struct {
	TasksCreated               int      `json:"tasks_created"`
	TasksCompleted             int      `json:"tasks_completed"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate"`
	TasksAverageCompletionTime *string  `json:"tasks_average_completion_time"`
}

func StatisticsDTOFromDomain(statistics domain.Statistics) StatisticsDTOResponse {
	var avgTime *string
	if statistics.TasksAverageCompletionTime != nil {
		duration := statistics.TasksAverageCompletionTime.String()
		avgTime = &duration
	}
	return StatisticsDTOResponse{
		TasksCreated:               statistics.TasksCreated,
		TasksCompleted:             statistics.TasksCompleted,
		TasksCompletedRate:         statistics.TasksCompletedRate,
		TasksAverageCompletionTime: avgTime,
	}
}
