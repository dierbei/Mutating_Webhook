delete:
	kubectl delete -f deploy/deployment.yaml
	kubectl delete -f deploy/tls-secret.yaml
	kubectl delete -f deploy/mutatingWebhookConfiguration.yaml
	kubectl delete -f deploy/cert-manager-1.5.3.yaml