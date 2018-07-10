.PHONY: force

polling: force
	vgo build -i -o ${GOPATH}/bin/fridgelethics-polling-service github.com/nafcollective/fridgelethics-service/polling/v0 