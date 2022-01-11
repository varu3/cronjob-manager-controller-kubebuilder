package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cronjobmanagerv1beta1 "github.com/varu3/cronjob-manager-kubebuilder/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CronjobManager controller", func() {
	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &cronjobmanagerv1beta1.CronJobManager{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.DeleteAllOf(ctx, &batchv1.CronJob{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		reconciler := CronJobManagerReconciler{
			Client:   mgr.GetClient(),
			Scheme:   mgr.GetScheme(),
			Recorder: mgr.GetEventRecorderFor("cronjobmanager-controller"),
		}

		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		stopFunc()
		time.Sleep(100 * time.Millisecond)
	})

	It("should create cronjob hoge-runner", func() {
		cjMng := newCronJobManager()
		err := k8sClient.Create(ctx, cjMng)
		Expect(err).NotTo(HaveOccurred())

		cj := batchv1.CronJob{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "hoge-runner"}, &cj)
		}).Should(Succeed())
		Expect(cj.Name).Should(Equal("hoge-runner"))
		Expect(cj.Spec.Schedule).Should(Equal("*/5 * * * *"))
		Expect(cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command).Should(Equal([]string{"echo", "hoge"}))
	})

	It("should be modified cronjob when update cronjobmanager", func() {
		cjMng := newCronJobManager()
		err := k8sClient.Create(ctx, cjMng)
		Expect(err).NotTo(HaveOccurred())

		cjMng.Spec.CronJobs[0].Name = "huga-runner"
		cjMng.Spec.CronJobs[0].Command = []string{"echo", "huga"}

		err = k8sClient.Update(ctx, cjMng)
		Expect(err).NotTo(HaveOccurred())

		cj := batchv1.CronJob{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "huga-runner"}, &cj)
		}).Should(Succeed())
		Expect(cj.Name).ShouldNot(Equal("hoge-runner"))
		Expect(cj.Name).Should(Equal("huga-runner"))
		Expect(cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command).Should(Equal([]string{"echo", "huga"}))
	})
})

func newCronJobManager() *cronjobmanagerv1beta1.CronJobManager {
	return &cronjobmanagerv1beta1.CronJobManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-cronjobmanager",
			Namespace: "test",
		},
		Spec: cronjobmanagerv1beta1.CronJobManagerSpec{
			Image: "debian:latest",
			CronJobs: []cronjobmanagerv1beta1.CronJobConfig{
				{
					Name:     "hoge-runner",
					Schedule: "*/5 * * * *",
					Command:  []string{"echo", "hoge"},
					Type:     "general",
				},
			},
		},
	}
}
