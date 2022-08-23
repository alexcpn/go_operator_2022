
# Creating a simple Kubernetes Operator with Kube-builder

Follow these two reference first

https://book.kubebuilder.io/quick-start.html

Second is bit complex  https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial.html

Read till Section 1.7 https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html

The rest you can follow this as the above is a non trivial example.

---


The easiest way to understand what a Kubernetes Operator can do is by building one. We will use the `Kubebuilder` frame-work to build one in Go language.  The other frame-work for this is `operator-sdk`.  Operator-SDK also uses Kubebuilder in the backend. We will create a simple operator that reads a custom CRD  that we create (read a Yaml of custom type similar to a Deployment yaml); and creates a Deployment out of that via code.

## Step 1 : Install Kubebuilder

Follow the make file to install Kubebuilder

`make once`

## Step 2: Init the project

`make init_project`

Note that we are giving DOMAIN and Project name as below in the make file. 

```
DOMAIN = mytest.io
PROJECT = testoperator
```

The Init will create a child folder  of name `PROJECT` and fill with Bolier plate code and files

## Step 3 : Create the API

`make create_crd`

Select `y` for both options Resources and Controller

```
cd testoperator && kubebuilder create api --group grpcapp --version v1 --kind Testoperartor && make manifests
Create Resource [y/n]
y
Create Controller [y/n]
y
```

This will create the CRD and Controller files. Out of the generated files three are impotant - The Controller, The Spec and the Yaml

```
testoperator_controller.go
testoperator_types.go
grpcapp_v1_testoperator.yaml
```
You can see all the generated files here in the two commits in this branch
https://github.com/alexcpn/go_operator_2022/compare/master...generated-code
## Step 4: Implement the logic

In this simple Operator we are going to read the CRD `testoperator/config/samples/grpcapp_v1_testoperartor.yaml` and create a deployment via code.

## Step 4.1
For this the minimum is  the Pod Image needed to create a deployment. We will add that to the above file

testoperator/config/samples/grpcapp_v1_testoperartor.yaml
```
# This is a sample Operator that will create a deployment with the name of the 
# podImage and also create a service with the given port and name
apiVersion: grpcapp.mytest.io/v1
kind: Testoperartor
metadata:
  name: testoperartor-sample
spec:
  # TODO(user): Add fields here
  # 1 ADDED
  podImage: alexcpn/run_server:1.2
  ```

## Step 4.2

Before we Apply this to the cluster we need to add the PodImage field to the controller types file `/testoperator/api/v1/testoperartor_types.go`
```
// TestoperartorSpec defines the desired state of Testoperartor
type TestoperartorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Testoperartor. Edit testoperartor_types.go to remove/update
	Foo string `json:"foo,omitempty"`

	// Le's create a service with this operator
	PodImage string `json:"podImage,omitempty"` //2 ADDED
}
```
Every time a new field is added, re-run the make file (Not in this folder, but in the child PROJECT folder, generated by Kubebuilder)

## Step 4.3

With the above step we will be able to successfully deploy the Yaml to the cluster with `make install`. But it will not be able to do anything

## Step 4.4

Add the controller logic

We first Get the Pod Image name from the deployed Kind (Step 4.3) and add the code to create a Deployment based on the retrieved Image name

```
func (r *TestoperartorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	//ADDED
	var testOperator grpcappv1.Testoperartor
	if err := r.Get(ctx, req.NamespacedName, &testOperator); err != nil {
		log.Log.Error(err, "unable to fetch Test Operator")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// ADDED - Block below
	log.Log.Info("Reconciling Test Operator", "Test Operator", testOperator)
	log.FromContext(ctx).Info("Pod Image is ", "PodImageName", testOperator.Spec.PodImage)
	// check if the PodImage is set
	if testOperator.Spec.PodImage == "" {
		log.Log.Info("Pod Image is not set")
	} else {
		log.Log.Info("Pod Image is set", "PodImageName", testOperator.Spec.PodImage)
	}
	//Lets create a deployment
	one := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testOperator.Name + "-deployment",
			Namespace: testOperator.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": testOperator.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": testOperator.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  testOperator.Name,
							Image: testOperator.Spec.PodImage,
						},
					},
				},
			},
		},
	}
	if err := r.Create(ctx, deployment); err != nil {
		log.Log.Error(err, "unable to create Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		return ctrl.Result{}, err
	}
	log.Log.Info("Created Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)

	return ctrl.Result{}, nil
}
```

To instruct the Kubebuilder to add RBAC for this operation we add the following too in the Reconcile function comments

```
// generate rbac to get,list, and watch pods // 3 ADDED
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// generate rbac to get, list, watch, create, update, patch, and delete deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
```

Also we do the following Imports


```
	appsv1 "k8s.io/api/apps/v1"                   //ADDED
	corev1 "k8s.io/api/core/v1"                   //ADDED
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" //ADDED
```    

## Step 4.5

Test it by deploying to cluster via `make run`


## Output

```
alex@pop-os:~/coding/app_fw/go_operator/testoperator$ make run
test -s /home/alex/coding/app_fw/go_operator/testoperator/bin/controller-gen || GOBIN=/home/alex/coding/app_fw/go_operator/testoperator/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2
/home/alex/coding/app_fw/go_operator/testoperator/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/alex/coding/app_fw/go_operator/testoperator/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./main.go
1.661231577844857e+09   INFO    controller-runtime.metrics      Metrics server is starting to listen    {"addr": ":8080"}
1.6612315778451552e+09  INFO    setup   starting manager
1.6612315778454158e+09  INFO    Starting server {"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
1.6612315778454406e+09  INFO    Starting server {"kind": "health probe", "addr": "[::]:8081"}
1.6612315778456118e+09  INFO    Starting EventSource    {"controller": "testoperartor", "controllerGroup": "grpcapp.mytest.io", "controllerKind": "Testoperartor", "source": "kind source: *v1.Testoperartor"}
1.661231577845636e+09   INFO    Starting Controller     {"controller": "testoperartor", "controllerGroup": "grpcapp.mytest.io", "controllerKind": "Testoperartor"}
1.66123157794669e+09    INFO    Starting workers        {"controller": "testoperartor", "controllerGroup": "grpcapp.mytest.io", "controllerKind": "Testoperartor", "worker count": 1}
1.6612315779469857e+09  INFO    Reconciling Test Operator       {"Test Operator": {"kind":"Testoperartor","apiVersion":"grpcapp.mytest.io/v1","metadata":{"name":"testoperartor-sample","namespace":"default","uid":"4c81b1d1-5e0e-42c3-a352-bce980542cd3","resourceVersion":"555018","generation":1,"creationTimestamp":"2022-08-22T12:10:26Z","annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"grpcapp.mytest.io/v1\",\"kind\":\"Testoperartor\",\"metadata\":{\"annotations\":{},\"name\":\"testoperartor-sample\",\"namespace\":\"default\"},\"spec\":{\"podImage\":\"alexcpn/run_server:1.2\"}}\n"},"managedFields":[{"manager":"kubectl-client-side-apply","operation":"Update","apiVersion":"grpcapp.mytest.io/v1","time":"2022-08-22T12:10:26Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:kubectl.kubernetes.io/last-applied-configuration":{}}},"f:spec":{".":{},"f:podImage":{}}}}]},"spec":{"podImage":"alexcpn/run_server:1.2"},"status":{}}}
1.6612315779471374e+09  INFO    Pod Image is    {"controller": "testoperartor", "controllerGroup": "grpcapp.mytest.io", "controllerKind": "Testoperartor", "testoperartor": {"name":"testoperartor-sample","namespace":"default"}, "namespace": "default", "name": "testoperartor-sample", "reconcileID": "dcfce095-8c44-4eb6-bb56-91ad52d550c9", "PodImageName": "alexcpn/run_server:1.2"}
1.6612315779471476e+09  INFO    Pod Image is set        {"PodImageName": "alexcpn/run_server:1.2"}
1.661231577953225e+09   INFO    Created Deployment      {"Deployment.Namespace": "default", "Deployment.Name": "testoperartor-sample-deployment"}
````

In the cluster

```
$ kubectl get deployment
NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
testoperartor-sample-deployment   1/1     1            1           31s

$ kubectl get pods
NAME                                               READY   STATUS    RESTARTS   AGE
testoperartor-sample-deployment-55645ff5cb-m4rjr   1/1     Running   0          44s

kubectl get testoperartor
NAME                   AGE
testoperartor-sample   17h

$ kubectl get testoperartor -o yaml
apiVersion: v1
items:
- apiVersion: grpcapp.mytest.io/v1
  kind: Testoperartor
  metadata:
    annotations:
                  ....
    creationTimestamp: "2022-08-22T12:10:26Z"
    generation: 1
    name: testoperartor-sample
    namespace: default
    resourceVersion: "555018"
    uid: 4c81b1d1-5e0e-42c3-a352-bce980542cd3
  spec:
    podImage: alexcpn/run_server:1.2
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

