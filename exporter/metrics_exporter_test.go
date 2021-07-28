package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	sdkTesting "github.com/brigadecore/brigade/sdk/v2/testing"
	authnTesting "github.com/brigadecore/brigade/sdk/v2/testing/authn"
	coreTesting "github.com/brigadecore/brigade/sdk/v2/testing/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsExporter(t *testing.T) {
	exporter := newMetricsExporter(
		&sdkTesting.MockAPIClient{
			CoreClient:  &coreTesting.MockAPIClient{},
			AuthnClient: &authnTesting.MockAPIClient{},
		},
		5*time.Second,
	)
	require.NotNil(t, exporter.coreClient)
	require.NotNil(t, exporter.authnClient)
	require.NotNil(t, exporter.scrapeInterval)
	require.NotNil(t, exporter.projectsGauge)
	require.NotNil(t, exporter.usersGauge)
	require.NotNil(t, exporter.serviceAccountsGauge)
	require.NotNil(t, exporter.allWorkersByPhase)
	require.NotNil(t, exporter.pendingJobsGauge)
}

func TestRecordProjectsCount(t *testing.T) {
	testCases := []struct {
		name       string
		exporter   *metricsExporter
		assertions func(*metricsExporter, error)
	}{
		{
			name: "error listing projects",
			exporter: &metricsExporter{
				coreClient: &coreTesting.MockAPIClient{
					ProjectsClient: &coreTesting.MockProjectsClient{
						ListFn: func(
							context.Context,
							*core.ProjectsSelector,
							*meta.ListOptions,
						) (core.ProjectList, error) {
							return core.ProjectList{}, errors.New("something went wrong")
						},
					},
				},
				projectsGauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.Error(t, err)
				require.Equal(t, "something went wrong", err.Error())
				assert.Equal(t, 0.0, testutil.ToFloat64(exporter.projectsGauge))
			},
		},
		{
			name: "success",
			exporter: &metricsExporter{
				coreClient: &coreTesting.MockAPIClient{
					ProjectsClient: &coreTesting.MockProjectsClient{
						ListFn: func(
							context.Context,
							*core.ProjectsSelector,
							*meta.ListOptions,
						) (core.ProjectList, error) {
							return core.ProjectList{
								ListMeta: meta.ListMeta{
									RemainingItemCount: 1,
								},
								Items: []core.Project{
									{}, // Return one project
								},
							}, nil
						},
					},
				},
				projectsGauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.NoError(t, err)
				assert.Equal(t, 2.0, testutil.ToFloat64(exporter.projectsGauge))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.exporter.recordProjectsCount()
			testCase.assertions(testCase.exporter, err)
		})
	}
}

func TestRecordUsersCount(t *testing.T) {
	testCases := []struct {
		name       string
		exporter   *metricsExporter
		assertions func(*metricsExporter, error)
	}{
		{
			name: "error listing users",
			exporter: &metricsExporter{
				authnClient: &authnTesting.MockAPIClient{
					UsersClient: &authnTesting.MockUsersClient{
						ListFn: func(
							context.Context,
							*authn.UsersSelector,
							*meta.ListOptions,
						) (authn.UserList, error) {
							return authn.UserList{}, errors.New("something went wrong")
						},
					},
				},
				usersGauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.Error(t, err)
				require.Equal(t, "something went wrong", err.Error())
				require.Equal(t, 0.0, testutil.ToFloat64(exporter.usersGauge))
			},
		},
		{
			name: "success",
			exporter: &metricsExporter{
				authnClient: &authnTesting.MockAPIClient{
					UsersClient: &authnTesting.MockUsersClient{
						ListFn: func(
							context.Context,
							*authn.UsersSelector,
							*meta.ListOptions,
						) (authn.UserList, error) {
							return authn.UserList{
								ListMeta: meta.ListMeta{
									RemainingItemCount: 1,
								},
								Items: []authn.User{
									{}, // Return one user
								},
							}, nil
						},
					},
				},
				usersGauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.NoError(t, err)
				require.Equal(t, 2.0, testutil.ToFloat64(exporter.usersGauge))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.exporter.recordUsersCount()
			testCase.assertions(testCase.exporter, err)
		})
	}
}

func TestRecordServiceAccountsCount(t *testing.T) {
	testCases := []struct {
		name       string
		exporter   *metricsExporter
		assertions func(*metricsExporter, error)
	}{
		{
			name: "error listing service accounts",
			exporter: &metricsExporter{
				authnClient: &authnTesting.MockAPIClient{
					ServiceAccountsClient: &authnTesting.MockServiceAccountsClient{
						ListFn: func(
							context.Context,
							*authn.ServiceAccountsSelector,
							*meta.ListOptions,
						) (authn.ServiceAccountList, error) {
							return authn.ServiceAccountList{},
								errors.New("something went wrong")
						},
					},
				},
				serviceAccountsGauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.Error(t, err)
				require.Equal(t, "something went wrong", err.Error())
				assert.Equal(t, 0.0, testutil.ToFloat64(exporter.serviceAccountsGauge))
			},
		},
		{
			name: "success",
			exporter: &metricsExporter{
				authnClient: &authnTesting.MockAPIClient{
					ServiceAccountsClient: &authnTesting.MockServiceAccountsClient{
						ListFn: func(
							context.Context,
							*authn.ServiceAccountsSelector,
							*meta.ListOptions,
						) (authn.ServiceAccountList, error) {
							return authn.ServiceAccountList{
								ListMeta: meta.ListMeta{
									RemainingItemCount: 1,
								},
								Items: []authn.ServiceAccount{
									{}, // Return 1 service account
								},
							}, nil
						},
					},
				},
				serviceAccountsGauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.NoError(t, err)
				assert.Equal(t, 2.0, testutil.ToFloat64(exporter.serviceAccountsGauge))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.exporter.recordServiceAccountsCount()
			testCase.assertions(testCase.exporter, err)
		})
	}
}

func TestRecordEventCountsByWorkersPhase(t *testing.T) {
	testCases := []struct {
		name       string
		exporter   *metricsExporter
		assertions func(*metricsExporter, error)
	}{
		{
			name: "error listing events",
			exporter: &metricsExporter{
				coreClient: &coreTesting.MockAPIClient{
					EventsClient: &coreTesting.MockEventsClient{
						ListFn: func(
							context.Context,
							*core.EventsSelector,
							*meta.ListOptions,
						) (core.EventList, error) {
							return core.EventList{}, errors.New("something went wrong")
						},
					},
				},
				allWorkersByPhase: prometheus.NewGaugeVec(
					prometheus.GaugeOpts{},
					[]string{"workerPhase"},
				),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.Error(t, err)
				require.Equal(t, "something went wrong", err.Error())
				for _, phase := range core.WorkerPhasesAll() {
					assert.Equal(
						t,
						0.0,
						testutil.ToFloat64(
							exporter.allWorkersByPhase.With(
								prometheus.Labels{"workerPhase": string(phase)}),
						),
					)
				}
			},
		},
		{
			name: "success",
			exporter: &metricsExporter{
				coreClient: &coreTesting.MockAPIClient{
					EventsClient: &coreTesting.MockEventsClient{
						ListFn: func(
							_ context.Context,
							selector *core.EventsSelector,
							_ *meta.ListOptions,
						) (core.EventList, error) {
							return core.EventList{
								Items: []core.Event{
									{
										Worker: &core.Worker{
											Status: core.WorkerStatus{
												Phase: selector.WorkerPhases[0],
											},
										},
									},
								},
							}, nil
						},
					},
				},
				allWorkersByPhase: prometheus.NewGaugeVec(
					prometheus.GaugeOpts{},
					[]string{"workerPhase"},
				),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.NoError(t, err)
				for _, phase := range core.WorkerPhasesAll() {
					assert.Equal(
						t,
						1.0,
						testutil.ToFloat64(
							exporter.allWorkersByPhase.With(
								prometheus.Labels{"workerPhase": string(phase)}),
						),
					)
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.exporter.recordEventCountsByWorkersPhase()
			testCase.assertions(testCase.exporter, err)
		})
	}
}

func TestRecordPendingJobsCount(t *testing.T) {
	testCases := []struct {
		name       string
		exporter   *metricsExporter
		assertions func(*metricsExporter, error)
	}{
		{
			name: "error listing events",
			exporter: &metricsExporter{
				coreClient: &coreTesting.MockAPIClient{
					EventsClient: &coreTesting.MockEventsClient{
						ListFn: func(
							context.Context,
							*core.EventsSelector,
							*meta.ListOptions,
						) (core.EventList, error) {
							return core.EventList{}, errors.New("something went wrong")
						},
					},
				},
				pendingJobsGauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.Error(t, err)
				require.Equal(t, "something went wrong", err.Error())
				assert.Equal(t, 0.0, testutil.ToFloat64(exporter.pendingJobsGauge))
			},
		},
		{
			name: "success",
			exporter: &metricsExporter{
				coreClient: &coreTesting.MockAPIClient{
					EventsClient: &coreTesting.MockEventsClient{
						ListFn: func(
							context.Context,
							*core.EventsSelector,
							*meta.ListOptions,
						) (core.EventList, error) {
							return core.EventList{
								Items: []core.Event{
									{
										Worker: &core.Worker{
											Jobs: []core.Job{
												{ // 1 pending job
													Status: &core.JobStatus{
														Phase: core.JobPhasePending,
													},
												},
											},
										},
									},
								},
							}, nil
						},
					},
				},
				pendingJobsGauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
			},
			assertions: func(exporter *metricsExporter, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1.0, testutil.ToFloat64(exporter.pendingJobsGauge))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.exporter.recordPendingJobsCount()
			testCase.assertions(testCase.exporter, err)
		})
	}
}
