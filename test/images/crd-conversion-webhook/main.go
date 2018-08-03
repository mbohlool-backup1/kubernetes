/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Config contains the server (the webhook) cert and key.
type Config struct {
	CertFile string
	KeyFile  string
}

func (c *Config) addFlags() {
	flag.StringVar(&c.CertFile, "tls-cert-file", c.CertFile, ""+
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated "+
		"after server cert).")
	flag.StringVar(&c.KeyFile, "tls-private-key-file", c.KeyFile, ""+
		"File containing the default x509 private key matching --tls-cert-file.")
}

func convertCRD(Object *unstructured.Unstructured, version string) (*unstructured.Unstructured, error) {
	glog.V(2).Info("converting crd")

	convertedObject := Object.DeepCopy()

	// convert it
	convertedObject.SetAPIVersion(version)

	return convertedObject, nil
}

type convertFunc func(Object *unstructured.Unstructured, version string) (*unstructured.Unstructured, error)

func serve(w http.ResponseWriter, r *http.Request, convert convertFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	glog.V(2).Info(fmt.Sprintf("handling request: %v", body))
	convertReview := v1beta1.ConversionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &convertReview); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		crd := unstructured.Unstructured{}
		if err := crd.UnmarshalJSON(convertReview.Request.Object.Raw); err != nil {
			glog.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		convertedCRD, err := convert(&crd, convertReview.Request.APIVersion)
		if err != nil {
			glog.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		convertReview.Response.ConvertedObject = runtime.RawExtension{
			Object: convertedCRD,
		}
	}
	glog.V(2).Info(fmt.Sprintf("sending response: %v", convertReview.Response))

	convertReview.Response.UID = convertReview.Request.UID
	// reset the request, it is not needed in a response.
	convertReview.Request = &v1beta1.ConversionRequest{}

	resp, err := json.Marshal(convertReview)
	if err != nil {
		glog.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
	}
}

func serveConvert(w http.ResponseWriter, r *http.Request) {
	serve(w, r, convertCRD)
}

func main() {
	var config Config
	config.addFlags()
	flag.Parse()

	http.HandleFunc("/convert", serveConvert)
	clientset := getClient()
	server := &http.Server{
		Addr:      ":443",
		TLSConfig: configTLS(config, clientset),
	}
	server.ListenAndServeTLS("", "")
}
