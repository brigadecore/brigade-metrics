package main

import (
	"context"
	"log"
	"time"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metricsExporter struct {
	coreClient           sdk.CoreClient
	authnClient          sdk.AuthnClient
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

func (m *metricsExporter) start(ctx context.Context) {
	go m.recordMetric(ctx, func() error {
		return m.recordProjectsCount()
	})
	go m.recordMetric(ctx, func() error {
		return m.recordUsersCount()
	})
	go m.recordMetric(ctx, func() error {
		return m.recordServiceAccountsCount()
	})
	go m.recordMetric(ctx, func() error {
		return m.recordEventCountsByWorkersPhase()
	})
	go m.recordMetric(ctx, func() error {
		return m.recordPendingJobsCount()
	})
}

func (m *metricsExporter) recordMetric(
	ctx context.Context,
	recordFn func() error,
) {
	ticker := time.NewTicker(m.scrapeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := recordFn(); err != nil {
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
		&sdk.ProjectsSelector{},
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
		&sdk.UsersSelector{},
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
		&sdk.ServiceAccountsSelector{},
		&meta.ListOptions{},
	)
	if err != nil {
		return err
	}
	m.serviceAccountsGauge.Set(
		float64(
			int64(len(serviceAccounts.Items)) +
				serviceAccounts.RemainingItemCount,
		),
	)
	return nil
}

func (m *metricsExporter) recordEventCountsByWorkersPhase() error {
	// brigade_events_by_worker_phase
	for _, phase := range sdk.WorkerPhasesAll() {
		events, err := m.coreClient.Events().List(
			context.Background(),
			&sdk.EventsSelector{
				WorkerPhases: []sdk.WorkerPhase{phase},
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
			&sdk.EventsSelector{
				WorkerPhases: []sdk.WorkerPhase{sdk.WorkerPhaseRunning},
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
				if job.Status.Phase == sdk.JobPhasePending {
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
