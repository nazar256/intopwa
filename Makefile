deploy: deploy-frontend deploy-backend

deploy-frontend:
	npx firebase deploy --only hosting

deploy-backend:
	cd worker && make deploy