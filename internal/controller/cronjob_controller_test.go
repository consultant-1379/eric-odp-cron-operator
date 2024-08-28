package controller

import (
	"context"
	"eric-odp-cron-operator/internal/fsclient"
	"regexp"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Test_cronJobReconciler(t *testing.T) {
	logger := zap.New(zap.WriteTo(ginkgo.GinkgoWriter))
	logf.SetLogger(logger)
	var menu map[string]string
	menu = map[string]string{
		"9 11 1,18 4 1":          "tar -zcf /var/backups/home.tgz /home/",
		"@midnight":              "/home/maverick/bin/cleanup-logs.sh",
		"11 11 1,10-16,18 4 2-5": "/home/ubuntu/command.sh > /tmp/output 2>&1",
	}

	tests := []struct {
		cronJobReconciler CronJobReconciler
		ctx               context.Context
		req               ctrl.Request
		expectedError     string
	}{
		{
			cronJobReconciler: CronJobReconciler{
				Client: nil,
			},
			ctx: context.TODO(),
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "test",
					Name:      "crontab",
				},
			},
			expectedError: "test",
		},
	}
	for _, test := range tests {
		odpuser := test.req.Name
		_, userSchedules := fsclient.GetFSEntryForUser(odpuser)
		for _, cronJobcommand := range userSchedules {
			re := regexp.MustCompile(cronregex)
			schedule := strings.Join(re.FindAllString(cronJobcommand, -1), "")
			command := strings.Split(cronJobcommand, schedule)[1]
			assert.Equal(t, strings.TrimSpace(command), menu[strings.TrimSpace(schedule)], "The two commands should be the same.")
		}

	}
}
