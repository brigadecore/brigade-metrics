package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	sdkTesting "github.com/brigadecore/brigade/sdk/v3/testing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsExporter(t *testing.T) {
	exporter := newMetricsExporter(
		&sdkTesting.MockAPIClient{
			CoreClient:  &sdkTesting.MockCoreClient{},
			AuthnClient: &sdkTesting.MockAuthnClient{},
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
				coreClient: &sdkTesting.MockCoreClient{
					ProjectsClient: &sdkTesting.MockProjectsClient{
						ListFn: func(
							context.Context,
							*sdk.ProjectsSelector,
							*meta.ListOptions,
						) (sdk.ProjectList, error) {
							return sdk.ProjectList{}, errors.New("something went wrong")
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
				coreClient: &sdkTesting.MockCoreClient{
					ProjectsClient: &sdkTesting.MockProjectsClient{
						ListFn: func(
							context.Context,
							*sdk.ProjectsSelector,
							*meta.ListOptions,
						) (sdk.ProjectList, error) {
							return sdk.ProjectList{
								ListMeta: meta.ListMeta{
									RemainingItemCount: 1,
								},
								Items: []sdk.Project{
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
				authnClient: &sdkTesting.MockAuthnClient{
					UsersClient: &sdkTesting.MockUsersClient{
						ListFn: func(
							context.Context,
							*sdk.UsersSelector,
							*meta.ListOptions,
						) (sdk.UserList, error) {
							return sdk.UserList{}, errors.New("something went wrong")
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
				authnClient: &sdkTesting.MockAuthnClient{
					UsersClient: &sdkTesting.MockUsersClient{
						ListFn: func(
							context.Context,
							*sdk.UsersSelector,
							*meta.ListOptions,
						) (sdk.UserList, error) {
							return sdk.UserList{
								ListMeta: meta.ListMeta{
									RemainingItemCount: 1,
								},
								Items: []sdk.User{
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
				authnClient: &sdkTesting.MockAuthnClient{
					ServiceAccountsClient: &sdkTesting.MockServiceAccountsClient{
						ListFn: func(
							context.Context,
							*sdk.ServiceAccountsSelector,
							*meta.ListOptions,
						) (sdk.ServiceAccountList, error) {
							return sdk.ServiceAccountList{},
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
				authnClient: &sdkTesting.MockAuthnClient{
					ServiceAccountsClient: &sdkTesting.MockServiceAccountsClient{
						ListFn: func(
							context.Context,
							*sdk.ServiceAccountsSelector,
							*meta.ListOptions,
						) (sdk.ServiceAccountList, error) {
							return sdk.ServiceAccountList{
								ListMeta: meta.ListMeta{
									RemainingItemCount: 1,
								},
								Items: []sdk.ServiceAccount{
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
				coreClient: &sdkTesting.MockCoreClient{
					EventsClient: &sdkTesting.MockEventsClient{
						ListFn: func(
							context.Context,
							*sdk.EventsSelector,
							*meta.ListOptions,
						) (sdk.EventList, error) {
							return sdk.EventList{}, errors.New("something went wrong")
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
				for _, phase := range sdk.WorkerPhasesAll() {
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
				coreClient: &sdkTesting.MockCoreClient{
					EventsClient: &sdkTesting.MockEventsClient{
						ListFn: func(
							_ context.Context,
							selector *sdk.EventsSelector,
							_ *meta.ListOptions,
						) (sdk.EventList, error) {
							return sdk.EventList{
								Items: []sdk.Event{
									{
										Worker: &sdk.Worker{
											Status: sdk.WorkerStatus{
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
				for _, phase := range sdk.WorkerPhasesAll() {
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
				coreClient: &sdkTesting.MockCoreClient{
					EventsClient: &sdkTesting.MockEventsClient{
						ListFn: func(
							context.Context,
							*sdk.EventsSelector,
							*meta.ListOptions,
						) (sdk.EventList, error) {
							return sdk.EventList{}, errors.New("something went wrong")
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
				coreClient: &sdkTesting.MockCoreClient{
					EventsClient: &sdkTesting.MockEventsClient{
						ListFn: func(
							context.Context,
							*sdk.EventsSelector,
							*meta.ListOptions,
						) (sdk.EventList, error) {
							return sdk.EventList{
								Items: []sdk.Event{
									{
										Worker: &sdk.Worker{
											Jobs: []sdk.Job{
												{ // 1 pending job
													Status: &sdk.JobStatus{
														Phase: sdk.JobPhasePending,
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
