/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"eric-odp-cron-operator/internal/fsclient"

	filewatcher "eric-odp-cron-operator/internal/watcher"

	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	fsScanTimeInMilliseconds = 5000
)

var (
	reconcileLog         = ctrl.Log.WithName("Reconciling")
	namespace            = os.Getenv("NAMESPACE")
	rootfspath           = os.Getenv("ROOT_FS_PATH")
	image                = os.Getenv("CRON_WRAPPER_IMAGE")
	timezone             = os.Getenv("TimeZone")
	imagePullSecret      = os.Getenv("IMAGE_PULL_SECRET")
	userlabel            = "com.ericsson.odp.username"
	odpcronjob           = "com.ericsson.odp.cronjob"
	cronUniqueIdentifier = "com.ericsson.odp.cron.unique.identifier"
	cronregex            = `^(\s*\S+\s){5}|(^\s*@\S+)`
)

type CronJobReconciler struct {
	client.Client
}

func (r *CronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconcileLog.Info("Reconcile :Starting reconciliation for " + req.NamespacedName.String())

	odpuser := req.Name

	associationsToCreate, userSchedules := fsclient.GetFSEntryForUser(odpuser)

	userk8sSchedules, jobstodelete := getValidK8sCronjobsSchedulesForUser(r, odpuser, ctx, userSchedules)

	reconcileLog.Info("Reconcile : ", "FS user schedules:", userSchedules)
	reconcileLog.Info("Reconcile :", "K8s user schedules:", userk8sSchedules)
	reconcileLog.Info("Reconcile :", "K8s Jobs to delete:", jobstodelete)

    for _, association := range associationsToCreate {
	   for identifier, cronJobcommand := range association.Usercrons {
		  if slices.Contains(userk8sSchedules, identifier) {
	         continue
		  }
		  err := createCronJob(r, cronJobcommand, odpuser, association.EnvVars)
		  if err != nil {
		      reconcileLog.Error(err, "Reconcile :Failed to create cron Job", "Cron user", odpuser, "Cron command", cronJobcommand)
          }
       }
	}
	for _, cronjob := range jobstodelete {
		reconcileLog.Info("Reconcile :", "Deleting Job ", cronjob.Annotations[cronUniqueIdentifier],
			"original job", cronjob.Name)
		err := r.Delete(ctx, cronjob)
		if err != nil {
			reconcileLog.Error(err, "Reconcile :", "Failed to delete cronjb", cronjob.Name, "With labels: ",
				cronjob.Annotations[cronUniqueIdentifier])
		}
	}
	return ctrl.Result{}, nil
}

func (r *CronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	labels := map[string]string{odpcronjob: "true"}
	go watcher(r)
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
		}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldGeneration := e.ObjectOld.GetGeneration()
				newGeneration := e.ObjectNew.GetGeneration()
				return oldGeneration != newGeneration
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				identifier := e.Object.GetAnnotations()[cronUniqueIdentifier]
				odpuser := e.Object.GetLabels()[userlabel]
				associations, _ := fsclient.GetFSEntryForUser(odpuser)
				reconcileLog.Info("SetupWithManager :", "In delete for ", odpuser, "Schedule ",
					e.Object.GetAnnotations()[cronUniqueIdentifier])
				for _, association := range associations {
					var mapOfEnvVars = association.EnvVars
					for usercronIdentifier, usercronCommand := range association.Usercrons{
					   if identifier == usercronIdentifier {
						  createCronJob(r, usercronCommand, odpuser, mapOfEnvVars)
					   }
					}
				}
				return false
			},
			CreateFunc: func(e event.CreateEvent) bool {
				identifier := e.Object.GetAnnotations()[cronUniqueIdentifier]
				odpuser := e.Object.GetLabels()[userlabel]
				_, userschedules := fsclient.GetFSEntryForUser(odpuser)
				_, exists := userschedules[identifier]
				if !exists {
					reconcileLog.Info("SetupWithManager :", "In Create for ", identifier, "userschedules", userschedules)
					err := r.Delete(context.Background(), e.Object)
					if err != nil {
						reconcileLog.Error(err, "Failed to delete job")
					}
				}
				return false
			},
		}).
		Complete(r)
}

func createCronJob(r *CronJobReconciler, cronJobcommand string, odpuser string, mapOfEnvVars map[string]string) error {

	re := regexp.MustCompile(cronregex)
	schedule := strings.Join(re.FindAllString(cronJobcommand, -1), "")
	command := strings.Split(cronJobcommand, schedule)[1]
    var mapAsString = fsclient.ConvertMapToString(mapOfEnvVars)
	identifier := fsclient.Md5calc(odpuser + cronJobcommand + image + mapAsString)
	uniqueIdentifier := uuid.NewString()
	hashusername := fsclient.Md5calc(odpuser + uniqueIdentifier)
	annotations := map[string]string{cronUniqueIdentifier: identifier}
	reconcileLog.Info("FS Reconcile:", "This cronjob needs to be created:", cronJobcommand, " for user ",
		odpuser, "with identifier " , identifier , " in ", namespace)
	backofflimit := int32(0)
	labels := map[string]string{odpcronjob: "true", userlabel: odpuser}
	job := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        "eric-odp-cron-" + hashusername,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					BackoffLimit: &backofflimit,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							ImagePullSecrets: []corev1.LocalObjectReference{
								{
									Name: imagePullSecret,
								},
							},
							Containers: []corev1.Container{
								{
									Image: image,
									Name:  "cron-wrapper",
									Env:   mergeEnvVariables(mapOfEnvVars, command, odpuser),
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}
	tzValue := SelectedTimezone(mapOfEnvVars)
	if tzValue != "" {
		job.Spec.TimeZone = &tzValue
	}
	err := r.Create(context.Background(), job)
	if err != nil {
		reconcileLog.Error(err, "createCronJob: Cound not create the job ")
		return err
	} else {
	   	reconcileLog.Info("FS Reconcile:", "Created:", cronJobcommand, " for user ",
       		odpuser, "with identifier " , identifier , " in ", namespace)
	}
	return nil
}

func getValidK8sCronjobsSchedulesForUser(r *CronJobReconciler, odpuser string, ctx context.Context, userSchedules map[string]string) ([]string, []*batchv1.CronJob) {
	labels := map[string]string{odpcronjob: "true", userlabel: odpuser}
	reconcileLog.V(1).Info("getValidK8sCronjobsSchedulesForUser: ", "Checking k8s for these user schedules ",
		userSchedules)
	var userk8sSchedules []string
	jobs := &batchv1.CronJobList{}
	var jobDeleteList []*batchv1.CronJob
	err := r.List(ctx, jobs, client.InNamespace(namespace), client.MatchingLabels(labels))
	if err != nil {
		return userk8sSchedules, jobDeleteList
	}
	for _, cronjob := range jobs.Items {
		identifier := cronjob.Annotations[cronUniqueIdentifier]
		if _, exists := userSchedules[identifier]; exists {
			userk8sSchedules = append(userk8sSchedules, identifier)
			continue
		}
		jobDeleteList = append(jobDeleteList, cronjob.DeepCopy())
	}
	return userk8sSchedules, jobDeleteList
}

func watcher(r *CronJobReconciler) {
	w := filewatcher.New()
	w.FilterOps(filewatcher.Create, filewatcher.Write)
	if err := w.AddRecursive(rootfspath); err != nil {
		log.Fatalln(err)
	}
	go func() {
		for {
			select {
			case watcherEvent := <-w.Event:
				reconcileLog.V(1).Info("Watcher ", " Event:", watcherEvent.Op, " path: ", watcherEvent.Path, "Name:",
					watcherEvent.Name(), "is Dir", watcherEvent.IsDir())
				if watcherEvent.Op == filewatcher.Write {
					if watcherEvent.IsDir() {
						reconcileLog.Info("Watcher : Call FS Reconcile:")
						_, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace,
							Name: watcherEvent.FileInfo.Name()}})
						if err != nil {
							reconcileLog.Info("Watcher :", " could not call reconcile", err)
						}
					}
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()
	for path, f := range w.WatchedFiles() {
		if f.IsDir() {
			continue
		}
		reconcileLog.Info("Watcher: Watching", "path", path, "file", f.Name())
		_, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: f.Name()}})
		if err != nil {
			reconcileLog.Error(err, "Watcher : Could not call reconcile")
		}
	}
	if err := w.Start(time.Millisecond * fsScanTimeInMilliseconds); err != nil {
		log.Fatalln(err)
	}
}

/*
TZ env variable takes precedence over the CRON_TZ
*/
func SelectedTimezone(myMap map[string]string) string {
	_, isTZ := myMap["TZ"]
	_, isCronTZ := myMap["CRON_TZ"]

	if isTZ {
		return myMap["TZ"]
	} else if isCronTZ {
		return myMap["CRON_TZ"]
	}
	return timezone
}

func mergeEnvVariables(mapOfEnvVars map[string]string, command, odpuser string) []corev1.EnvVar {
	res := make([]corev1.EnvVar, 0)

	for key, val := range mapOfEnvVars {
		if len(val) != 0 {
			res = append(res, corev1.EnvVar{
				Name:  key,
				Value: val,
			})
		}
	}

	res = append(res, corev1.EnvVar{
		Name:  "ODP-CRON-POD-CMD",
		Value: command,
	})

	res = append(res, corev1.EnvVar{
		Name:  "ODP-CRON-POD-USERNAME",
		Value: odpuser,
	})
	return res
}

