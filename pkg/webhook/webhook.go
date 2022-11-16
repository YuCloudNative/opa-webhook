package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"opa-webhook/pkg/constants"
	"opa-webhook/pkg/patch"
)

// OpaWebhookServer for opa webhook parameter
type OpaWebhookServer struct {
	Server *http.Server
}

// OpaWebHookServerParameters opa webhook Server parameters
type OpaWebHookServerParameters struct {
	// webhook server Port
	Port int
	// path to the x509 certificate for https
	CertFile string
	// path to the x509 private key matching `CertFile`
	KeyFile string
}

func (opaWebhookServer *OpaWebhookServer) admissionReviewFromRequest(r *http.Request) (*admissionv1.AdmissionReview, error) {
	// Validate that the incoming content type is correct.
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("expected application/json content-type")
	}

	admissionReviewRequest := &admissionv1.AdmissionReview{}
	err := json.NewDecoder(r.Body).Decode(&admissionReviewRequest)
	if err != nil {
		return nil, err
	}
	return admissionReviewRequest, nil
}

func (opaWebhookServer *OpaWebhookServer) admissionResponseFromReview(admReview *admissionv1.AdmissionReview) (*admissionv1.AdmissionResponse, error) {
	admissionResponse := &admissionv1.AdmissionResponse{}

	// Decode the pod from the AdmissionReview.
	req := admReview.Request
	rawRequest := req.Object.Raw
	pod := &corev1.Pod{}

	err := json.NewDecoder(bytes.NewReader(rawRequest)).Decode(pod)
	if err != nil {
		err := fmt.Errorf("error decoding raw pod: %v", err)
		return nil, err
	}

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	podAnnotations := pod.GetAnnotations()
	if podAnnotations == nil {
		podAnnotations = map[string]string{}
	}

	csmOpaInjectionAnnotation, ok := podAnnotations[constants.CsmOpaInjectionAnnotation]
	if ok && constants.CsmOpaInjectionAnnotationFalse == csmOpaInjectionAnnotation {
		info := fmt.Sprintf("Skipping mutation for annotation %v=false", constants.CsmOpaInjectionAnnotation)
		glog.Infof(info)
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}, nil
	}

	metadata := &pod.ObjectMeta
	for _, namespace := range constants.IgnoredNamespaces {
		if metadata.Namespace == namespace {
			glog.Infof("Skipping mutation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return &admissionv1.AdmissionResponse{
				Allowed: true,
			}, nil
		}
	}

	// determine whether to perform mutation
	ok, err = opaWebhookServer.mutationRequired(pod)
	if !ok || err != nil {
		info := "injecting opa sidecar failed due to policy check mutationRequired rejected"
		glog.Errorf(info)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}, err
	}

	patchBytes, err := patch.CreatePatch(pod)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}, err
	}

	glog.Infof("patching the pod with: %s", string(patchBytes))
	if patchBytes != nil {
		admissionResponse.Allowed = true
		patchType := admissionv1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
		admissionResponse.Patch = patchBytes
	}

	return admissionResponse, nil
}

// HandleMutate method for webhook Server
func (opaWebhookServer *OpaWebhookServer) HandleMutate(writer http.ResponseWriter, request *http.Request) {
	glog.Infof("OPA Webhook Server start handleMutate for request")
	admReview, err := opaWebhookServer.admissionReviewFromRequest(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		glog.Errorf("error getting admission review from request: %v", err)
		return
	}

	admResp, err := opaWebhookServer.admissionResponseFromReview(admReview)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		glog.Errorf("error getting admissionResponseFromReview review from request: %v", err)
		return
	}

	// the final response will be another admission review
	var admissionReviewResponse admissionv1.AdmissionReview
	admissionReviewResponse.Response = admResp
	admissionReviewResponse.SetGroupVersionKind(admReview.GroupVersionKind())
	admissionReviewResponse.Response.UID = admReview.Request.UID

	resp, err := json.Marshal(admissionReviewResponse)
	if err != nil {
		msg := fmt.Sprintf("error marshaling response: %v", err)
		glog.Errorf(msg)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	glog.Infof("allowing pod %v", string(resp))
	writer.Write(resp)
}

// HandleRoot method for webhook Server
func (opaWebhookServer *OpaWebhookServer) HandleRoot(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "hello %q", html.EscapeString(request.URL.Path))
}

// mutationRequired check whether the target resource need to be mutated
func (opaWebhookServer *OpaWebhookServer) mutationRequired(pod *corev1.Pod) (bool, error) {
	metadata := &pod.ObjectMeta

	podLabels := metadata.GetLabels()
	if podLabels == nil {
		podLabels = map[string]string{}
	}
	glog.Infof("pod has following labels %v", podLabels)

	podAnnotations := metadata.GetAnnotations()
	if podAnnotations == nil {
		podAnnotations = map[string]string{}
	}
	glog.Infof("pod has following annotations %v", podAnnotations)

	regoSourceType, ok := podAnnotations[constants.RegoSourceType]
	if !ok {
		info := fmt.Sprintf("Skipping mutation for lacking annotation %s ", constants.RegoSourceType)
		glog.Infof(info)
		return false, errors.New(info)
	}

	if regoSourceType == "" {
		info := fmt.Sprintf("Skipping mutation for annotation %s is null", constants.RegoSourceType)
		glog.Infof(info)
		return false, errors.New(info)
	}

	if _, ok = podAnnotations[constants.OpaImageAddr]; !ok {
		info := fmt.Sprintf("Skipping mutation for lacking annotation %s ", constants.OpaImageAddr)
		glog.Infof(info)
		return false, errors.New(info)
	}

	glog.Infof("opa webhook server mutation policy strategy is %v", regoSourceType)

	if constants.Configmap == regoSourceType {
		if _, ok = podAnnotations[constants.RegoConfigMapName]; !ok {
			info := fmt.Sprintf("Skipping mutation for annotation %s is null", constants.RegoConfigMapName)
			glog.Infof(info)
			return false, errors.New(info)
		}
		if _, ok = podAnnotations[constants.RegoConfigMapNameKey]; !ok {
			info := fmt.Sprintf("Skipping mutation for annotation %s is null", constants.RegoConfigMapNameKey)
			glog.Infof(info)
			return false, errors.New(info)
		}
		if _, ok = podAnnotations[constants.RegoConfigYamlGrpcPolicyPath]; !ok {
			info := fmt.Sprintf("Skipping mutation for annotation %s is null", constants.RegoConfigYamlGrpcPolicyPath)
			glog.Infof(info)
			return false, errors.New(info)
		}
	} else if constants.Bundle == regoSourceType {
		if _, ok = podAnnotations[constants.RegoConfigYamlGrpcPolicyPath]; !ok {
			info := fmt.Sprintf("Skipping mutation for annotation %s is null", constants.RegoConfigYamlGrpcPolicyPath)
			glog.Infof(info)
			return false, errors.New(info)
		}
	} else {
		return false, errors.New(fmt.Sprintf("strategy type %s is not supported", regoSourceType))
	}

	return true, nil
}
