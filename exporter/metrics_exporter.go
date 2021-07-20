package main

import (
	"context"
	"log"
	"time"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metricsExporter struct {
	apiClient            sdk.APIClient
	scrapeInterval       time.Duration
	totalProjects        prometheus.Gauge
	totalUsers           prometheus.Gauge
	totalServiceAccounts prometheus.Gauge
	allWorkersByPhase    *prometheus.GaugeVec
	totalPendingJobs     prometheus.Gauge
}

func newMetricsExporter(
	apiClient sdk.APIClient,
	scrapeInterval time.Duration,
) *metricsExporter {
	return &metricsExporter{
		apiClient:      apiClient,
		scrapeInterval: scrapeInterval,
		totalProjects: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "brigade_projects_total",
				Help: "The total number of brigade projects",
			},
		),
		totalUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "brigade_users_total",
				Help: "The total number of users",
			},
		),
		totalServiceAccounts: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "brigade_service_accounts_total",
				Help: "The total number of service accounts",
			},
		),
		allWorkersByPhase: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "brigade_all_workers_by_phase",
				Help: "All workers separated by phase",
			},
			[]string{"workerPhase"},
		),
		totalPendingJobs: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "brigade_pending_jobs_total",
				Help: "The total number of pending jobs",
			},
		),
	}
}

func (m *metricsExporter) run(ctx context.Context) {
	ticker := time.NewTicker(m.scrapeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.recordMetrics()
		case <-ctx.Done():
			return
		}
	}
}

func (m *metricsExporter) recordMetrics() {
	// brigade_projects_total
	projects, err := m.apiClient.Core().Projects().List(
		context.Background(),
		&core.ProjectsSelector{},
		&meta.ListOptions{},
	)
	if err != nil {
		log.Println(err)
	} else {
		m.totalProjects.Set(float64(len(projects.Items) +
			int(projects.RemainingItemCount)))
	}

	// brigade_users_total
	users, err := m.apiClient.Authn().Users().List(
		context.Background(),
		&authn.UsersSelector{},
		&meta.ListOptions{},
	)
	if err != nil {
		log.Println(err)
	} else {
		m.totalUsers.Set(float64(len(users.Items) +
			int(users.RemainingItemCount)))
	}

	// brigade_service_accounts_total
	serviceAccounts, err := m.apiClient.Authn().ServiceAccounts().List(
		context.Background(),
		&authn.ServiceAccountsSelector{},
		&meta.ListOptions{},
	)
	if err != nil {
		log.Println(err)
	} else {
		m.totalServiceAccounts.Set(
			float64(
				len(serviceAccounts.Items) +
					int(serviceAccounts.RemainingItemCount),
			),
		)
	}

	// brigade_all_workers_by_phase
	for _, phase := range core.WorkerPhasesAll() {
		var events core.EventList
		events, err = m.apiClient.Core().Events().List(
			context.Background(),
			&core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{phase},
			},
			&meta.ListOptions{},
		)
		if err != nil {
			log.Println(err)
		} else {
			m.allWorkersByPhase.With(
				prometheus.Labels{"workerPhase": string(phase)},
			).Set(float64(len(events.Items) + int(events.RemainingItemCount)))
		}

		// brigade_pending_jobs_total
		//
		// There is no way to query the API directly for pending Jobs, but only
		// running Workers should ever HAVE pending Jobs, so if we're currently
		// counting running Workers, we can iterate over those to count pending
		// jobs. Note, there's a cap on the max number of workers that can run
		// concurrently, so we assume that as long as that cap isn't enormous (which
		// would only occur on an enormous cluster), it's practical to iterate over
		// all the running workers.
		if phase == core.WorkerPhaseRunning {
			var pendingJobs int
			for {
				for _, event := range events.Items {
					for _, job := range event.Worker.Jobs {
						if job.Status.Phase == core.JobPhasePending {
							pendingJobs++
						}
					}
				}
				if events.Continue == "" {
					break
				}
				if events, err = m.apiClient.Core().Events().List(
					context.Background(),
					&core.EventsSelector{
						WorkerPhases: []core.WorkerPhase{phase},
					},
					&meta.ListOptions{Continue: events.Continue},
				); err != nil {
					log.Println(err)
					break
				}
			}
			if err == nil {
				m.totalPendingJobs.Set(float64(pendingJobs))
			}
		}
	}

}
