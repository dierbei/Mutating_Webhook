package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"

	admission "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tlsKeyName  = "tls.key"
	tlsCertName = "tls.crt"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/mutate", Mutation)

	certDir := os.Getenv("CERT_DIR")

	if err := http.ListenAndServeTLS(":8080", fmt.Sprintf("%s/%s", certDir, tlsCertName), fmt.Sprintf("%s/%s", certDir, tlsKeyName), r); err != nil {
		panic(err)
	}
}

func Mutation(ctx *gin.Context) {
	// 从 Context 对象中获取请求和响应信息
	r := ctx.Request
	w := ctx.Writer

	// 从 req 从去除请求体
	ar := new(admission.AdmissionReview)
	if err := json.NewDecoder(r.Body).Decode(ar); err != nil {
		handleError(w, nil, err)
		return
	}

	// 反序列化为目标对象
	pod := &corev1.Pod{}
	if err := json.Unmarshal(ar.Request.Object.Raw, pod); err != nil {
		handleError(w, ar, err)
		return
	}

	// 增加环境变量
	for i := 0; i < len(pod.Spec.Containers); i++ {
		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, corev1.EnvVar{
			Name:  "DEBUG",
			Value: "true",
		})
	}

	// 序列化更新之后的数据
	containersBytes, err := json.Marshal(&pod.Spec.Containers)
	if err != nil {
		handleError(w, ar, err)
		return
	}

	// 需要执行的操作
	patch := []JSONPatchEntry{
		{
			OP:    "replace",
			Path:  "/spec/containers",
			Value: containersBytes,
		},
	}
	patchBytes, err := json.Marshal(&patch)
	if err != nil {
		handleError(w, ar, err)
		return
	}

	// 构造返回值
	patchType := admission.PatchTypeJSONPatch
	response := &admission.AdmissionResponse{
		UID:       ar.Request.UID,
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}
	responseAR := &admission.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	}

	// 把结果写入 write
	json.NewEncoder(w).Encode(responseAR)
}

type JSONPatchEntry struct {
	OP    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value,omitempty"`
}

func handleError(w http.ResponseWriter, ar *admission.AdmissionReview, err error) {
	if err != nil {
		log.Println("[Error]", err.Error())
	}
	response := &admission.AdmissionResponse{
		Allowed: false,
	}
	if ar != nil {
		response.UID = ar.Request.UID
	}
	ar.Response = response
	json.NewEncoder(w).Encode(ar)
}
