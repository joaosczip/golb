package config

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/joaosczip/go-lb/internal/algorithms"
	"github.com/joaosczip/go-lb/internal/proxy"
	"github.com/joaosczip/go-lb/pkg/lb/targetgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FileReaderMock struct {
	mock.Mock
}

func (m *FileReaderMock) Read(path string) ([]byte, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

type ProxyFactoryMock struct {
	mock.Mock
}

func (m *ProxyFactoryMock) Create(host string, port int) proxy.Proxy {
	args := m.Called(host, port)
	return args.Get(0).(proxy.Proxy)
}

type TestSetup struct {
	fileReader *FileReaderMock
	proxyFactory *ProxyFactoryMock
	httpClient http.Client
	configLoader ConfigLoader
}

func setup() TestSetup {
	fileReader := new(FileReaderMock)
	proxyFactory := new(ProxyFactoryMock)
	httpClient := new(http.Client)
	configLoader := NewConfigLoader("config.yaml", httpClient, proxyFactory, fileReader)

	return TestSetup{
		fileReader: fileReader,
		proxyFactory: proxyFactory,
		httpClient: *httpClient,
		configLoader: *configLoader,
	}
}

func TestConfigLoader_Load(t *testing.T) {
	t.Run("Should return an error when file reader cannot read the file", func(t *testing.T) {
		testSetup := setup()
		
		testSetup.fileReader.On("Read", "config.yaml").Return([]byte(nil), errors.New("unexpected error"))

		_, err := testSetup.configLoader.Load()

		assert.EqualError(t, err, "could not read config file: unexpected error")
		testSetup.fileReader.AssertExpectations(t)
	})

	t.Run("Should return an error when the file content is invalid", func(t *testing.T) {
		testSetup := setup()

		testSetup.fileReader.On("Read", "config.yaml").Return([]byte("invalid content"), nil)

		_, err := testSetup.configLoader.Load()

		assert.NotNil(t, err)
		testSetup.fileReader.AssertExpectations(t)
	})

	t.Run("Should return a list of target groups on success", func(t *testing.T) {
		testSetup := setup()
		
		mockedYamlConfig, err := os.ReadFile("../../test/fixtures/mocked_config_file.yaml")

		assert.NoError(t, err)

		testSetup.fileReader.On("Read", "config.yaml").Return(mockedYamlConfig, nil)

		targetGroups, err := testSetup.configLoader.Load()

		assert.Nil(t, err)
		assert.Len(t, targetGroups, 1)
		assert.Len(t, targetGroups[0].Targets, 2)
		assert.Equal(t, targetGroups[0].Algorithm, algorithms.NewRoundRobin(targetGroups[0].Targets, testSetup.proxyFactory))

		assert.Equal(t, targetGroups[0].Targets, []*targetgroup.Target{
			{Host: "localhost", Port: 8080},
			{Host: "localhost", Port: 8081},
		})

		assert.Equal(t, targetGroups[0].HealthCheckConfig, &targetgroup.HealthCheckConfig{
			Interval: 1,
			Timeout: 2,
			FailureThreshold: 3,
			HealthyThreshold: 4,
			Path: "/health",
			HttpClient: &testSetup.httpClient,
		})
	})
}