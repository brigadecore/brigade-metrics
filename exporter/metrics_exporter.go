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
	coreClient           core.APIClient
	authnClient          authn.APIClient
	scrapeInterval       time.Duration
	projectsGauge        prometheus.Gauge
	usersGauge           prometheus.Gauge
	serviceAccountsGauge prometheus.Gauge
	allWorkersByPhase    *prometheus.GaugeVec
	pendingJobsGauge     prometheus.Gauge
}

func newMetricsExporter(
	apiClient sdk.APIClient,
	scrapeInterval time.Duration,
) *metricsExporter {
	return &metricsExporter{
		coreClient:     apiClient.Core(),
		authnClient:    apiClient.Authn(),
		scrapeInterval: scrapeInterval,
		projectsGauge: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "brigade_projects_total",
				Help: "The total number of projects",
			},
		),
		usersGauge: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "brigade_users_total",
				Help: "The total number of users",
			},
		),
		serviceAccountsGauge: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "brigade_service_accounts_total",
				Help: "The total number of service accounts",
			},
		),
		allWorkersByPhase: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "brigade_events_by_worker_phase",
				Help: "The total number of events grouped by worker phase",
			},
			[]string{"workerPhase"},
		),
		pendingJobsGauge: promauto.NewGauge(
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
			if err := m.recordProjectsCount(); err != nil {
				log.Println(err)
			}
			if err := m.recordUsersCount(); err != nil {
				log.Println(err)
			}
			if err := m.recordServiceAccountsCount(); err != nil {
				log.Println(err)
			}
			if err := m.recordEventCountsByWorkersPhase(); err != nil {
				log.Println(err)
			}
			if err := m.recordPendingJobsCount(); err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *metricsExporter) recordProjectsCount() error {
	// brigade_projects_total
	projects, err := m.coreClient.Projects().List(
		context.Background(),
		&core.ProjectsSelector{},
		&meta.ListOptions{},
	)
	if err != nil {
		return err
	}
	m.projectsGauge.Set(
		float64(len(projects.Items) + int(projects.RemainingItemCount)),
	)
	return nil
}

func (m *metricsExporter) recordUsersCount() error {
	// brigade_users_total
	users, err := m.authnClient.Users().List(
		context.Background(),
		&authn.UsersSelector{},
		&meta.ListOptions{},
	)
	if err != nil {
		return err
	}
	m.usersGauge.Set(
		float64(int64(len(users.Items)) + users.RemainingItemCount),
	)
	return nil
}

func (m *metricsExporter) recordServiceAccountsCount() error {
	// brigade_service_accounts_total
	serviceAccounts, err := m.authnClient.ServiceAccounts().List(
		context.Background(),
		&authn.ServiceAccountsSelector{},
		&meta.ListOptions{},
	)
	if err != nil {
		return err
	}
	m.serviceAccountsGauge.Set(
		float64(
			int64(len(serviceAccounts.Items)) + serviceAccounts.RemainingItemCount,
		),
	)
	return nil
}

func (m *metricsExporter) recordEventCountsByWorkersPhase() error {
	// brigade_events_by_worker_phase
	for _, phase := range core.WorkerPhasesAll() {
		events, err := m.coreClient.Events().List(
			context.Background(),
			&core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{phase},
			},
			&meta.ListOptions{},
		)
		if err != nil {
			return err
		}
		m.allWorkersByPhase.With(
			prometheus.Labels{"workerPhase": string(phase)},
		).Set(float64(len(events.Items) + int(events.RemainingItemCount)))
	}
	return nil
}

func (m *metricsExporter) recordPendingJobsCount() error {
	// brigade_pending_jobs_total
	var pendingJobs int
	var continueValue string
	for {
		events, err := m.coreClient.Events().List(
			context.Background(),
			&core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{core.WorkerPhaseRunning},
			},
			&meta.ListOptions{
				Continue: continueValue,
			},
		)
		if err != nil {
			return err
		}
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
		continueValue = events.Continue
	}
	m.pendingJobsGauge.Set(float64(pendingJobs))
	return nil
}
