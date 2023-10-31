package config

import (
	"crypto/tls"
	"time"

	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	flagMetricsBindAddr             = "metrics-bind-addr"
	flagHealthProbeBindAddr         = "health-probe-bind-addr"
	flagWebhookBindPort             = "webhook-bind-port"
	flagEnableLeaderElection        = "enable-leader-election"
	flagLeaderElectionID            = "leader-election-id"
	flagLeaderElectionNamespace     = "leader-election-namespace"
	flagLeaderElectionLeaseDuration = "leader-election-lease-duration"
	flagLeaderElectionRenewDeadline = "leader-election-renew-deadline"
	flagLeaderElectionRetryPeriod   = "leader-election-retry-period"
	flagWatchNamespace              = "watch-namespace"
	flagSyncPeriod                  = "sync-period"
	flagKubeconfig                  = "kubeconfig"
	flagWebhookCertDir              = "webhook-cert-dir"
	flagWebhookCertName             = "webhook-cert-file"
	flagWebhookKeyName              = "webhook-key-file"

	defaultKubeconfig                  = ""
	defaultLeaderElectionID            = "aws-load-balancer-controller-leader"
	defaultLeaderElectionNamespace     = ""
	defaultLeaderElectionLeaseDuration = 15 * time.Second
	defaultLeaderElectionRenewDeadline = 10 * time.Second
	defaultLeaderElectionRetryPeriod   = 2 * time.Second
	defaultWatchNamespace              = corev1.NamespaceAll
	defaultMetricsAddr                 = ":8080"
	defaultHealthProbeBindAddress      = ":61779"
	defaultSyncPeriod                  = 10 * time.Hour
	defaultWebhookBindPort             = 9443
	// High enough QPS to fit all expected use cases. QPS=0 is not set here, because
	// client code is overriding it.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases. Burst=0 is not set here, because
	// client code is overriding it.
	defaultBurst           = 1e6
	defaultWebhookCertDir  = ""
	defaultWebhookCertName = ""
	defaultWebhookKeyName  = ""
)

// RuntimeConfig stores the configuration for the controller-runtime
type RuntimeConfig struct {
	APIServer                   string
	KubeConfig                  string
	WebhookBindPort             int
	MetricsBindAddress          string
	HealthProbeBindAddress      string
	EnableLeaderElection        bool
	LeaderElectionID            string
	LeaderElectionNamespace     string
	LeaderElectionLeaseDuration time.Duration
	LeaderElectionRenewDeadline time.Duration
	LeaderElectionRetryPeriod   time.Duration
	WatchNamespace              string
	SyncPeriod                  time.Duration
	WebhookCertDir              string
	WebhookCertName             string
	WebhookKeyName              string
}

// BindFlags binds the command line flags to the fields in the config object
func (c *RuntimeConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.KubeConfig, flagKubeconfig, defaultKubeconfig,
		"Path to the kubeconfig file containing authorization and API server information.")
	fs.StringVar(&c.MetricsBindAddress, flagMetricsBindAddr, defaultMetricsAddr,
		"The address the metric endpoint binds to.")
	fs.StringVar(&c.HealthProbeBindAddress, flagHealthProbeBindAddr, defaultHealthProbeBindAddress,
		"The address the health probes binds to.")
	fs.IntVar(&c.WebhookBindPort, flagWebhookBindPort, defaultWebhookBindPort,
		"The TCP port the Webhook server binds to.")
	fs.BoolVar(&c.EnableLeaderElection, flagEnableLeaderElection, true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&c.LeaderElectionID, flagLeaderElectionID, defaultLeaderElectionID,
		"Name of the leader election ID to use for this controller")
	fs.StringVar(&c.LeaderElectionNamespace, flagLeaderElectionNamespace, defaultLeaderElectionNamespace,
		"Name of the leader election ID to use for this controller")
	fs.DurationVar(&c.LeaderElectionLeaseDuration, flagLeaderElectionLeaseDuration, defaultLeaderElectionLeaseDuration,
		"The duration that non-leader candidates will wait to force acquire leadership.")
	fs.DurationVar(&c.LeaderElectionRenewDeadline, flagLeaderElectionRenewDeadline, defaultLeaderElectionRenewDeadline,
		"The duration that the acting controlplane will retry refreshing leadership before giving up.")
	fs.DurationVar(&c.LeaderElectionRetryPeriod, flagLeaderElectionRetryPeriod, defaultLeaderElectionRetryPeriod,
		"The duration the LeaderElector clients should wait between tries of actions.")
	fs.StringVar(&c.WatchNamespace, flagWatchNamespace, defaultWatchNamespace,
		"Namespace the controller watches for updates to Kubernetes objects, If empty, all namespaces are watched.")
	fs.DurationVar(&c.SyncPeriod, flagSyncPeriod, defaultSyncPeriod,
		"Period at which the controller forces the repopulation of its local object stores.")
	fs.StringVar(&c.WebhookCertDir, flagWebhookCertDir, defaultWebhookCertDir, "WebhookCertDir is the directory that contains the webhook server key and certificate.")
	fs.StringVar(&c.WebhookCertName, flagWebhookCertName, defaultWebhookCertName, "WebhookCertName is the webhook server certificate name.")
	fs.StringVar(&c.WebhookKeyName, flagWebhookKeyName, defaultWebhookKeyName, "WebhookKeyName is the webhook server key name.")

}

// BuildRestConfig builds the REST config for the controller runtime
func BuildRestConfig(rtCfg RuntimeConfig) (*rest.Config, error) {
	var restCFG *rest.Config
	var err error
	if rtCfg.KubeConfig == "" {
		restCFG, err = rest.InClusterConfig()
	} else {
		restCFG, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: rtCfg.KubeConfig}, &clientcmd.ConfigOverrides{}).ClientConfig()
	}
	if err != nil {
		return nil, err
	}

	restCFG.QPS = defaultQPS
	restCFG.Burst = defaultBurst
	return restCFG, nil
}

// BuildRuntimeOptions builds the options for the controller runtime based on config
func BuildRuntimeOptions(rtCfg RuntimeConfig, scheme *runtime.Scheme) ctrl.Options {
	return ctrl.Options{
		Scheme:                     scheme,
		Port:                       rtCfg.WebhookBindPort,
		CertDir:                    rtCfg.WebhookCertDir,
		MetricsBindAddress:         rtCfg.MetricsBindAddress,
		HealthProbeBindAddress:     rtCfg.HealthProbeBindAddress,
		LeaderElection:             rtCfg.EnableLeaderElection,
		LeaderElectionResourceLock: resourcelock.ConfigMapsLeasesResourceLock,
		LeaderElectionID:           rtCfg.LeaderElectionID,
		LeaderElectionNamespace:    rtCfg.LeaderElectionNamespace,
		LeaseDuration:              &rtCfg.LeaderElectionLeaseDuration,
		RenewDeadline:              &rtCfg.LeaderElectionRenewDeadline,
		RetryPeriod:                &rtCfg.LeaderElectionRetryPeriod,
		Namespace:                  rtCfg.WatchNamespace,
		SyncPeriod:                 &rtCfg.SyncPeriod,
		ClientDisableCacheFor:      []client.Object{&corev1.Secret{}},
	}
}

// ConfigureWebhookServer set up the server cert for the webhook server.
func ConfigureWebhookServer(rtCfg RuntimeConfig, mgr ctrl.Manager) {
	mgr.GetWebhookServer().CertName = rtCfg.WebhookCertName
	mgr.GetWebhookServer().KeyName = rtCfg.WebhookKeyName
	mgr.GetWebhookServer().TLSOpts = []func(config *tls.Config){
		func(config *tls.Config) {
			config.MinVersion = tls.VersionTLS12
			config.CipherSuites = []uint16{
				// AEADs w/ ECDHE
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,

				// AEADs w/o ECDHE
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			}
		},
	}
}
