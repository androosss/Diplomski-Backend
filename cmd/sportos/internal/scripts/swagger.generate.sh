SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -c "backoffice/v1" -c "internal/docs/docs_bo" -c "api/dto" -o ./internal/swagger/swagger_bo.json

SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -c "public/v1" -c "internal/docs/docs_pub" -c "api/dto" -o ./internal/swagger/swagger_pub.json

SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -c "pp/v1" -c "internal/docs/docs_pp" -c "api/dto" -o ./internal/swagger/swagger_pp.json

swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_bo.json ./internal/swagger/swagger_bo.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_bo.json ./internal/swagger/swagger_bo.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_bo.json ./internal/swagger/swagger_bo.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_bo.json ./internal/swagger/swagger_bo.json

swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_pp.json ./internal/swagger/swagger_pp.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_pp.json ./internal/swagger/swagger_pp.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_pp.json ./internal/swagger/swagger_pp.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_pp.json ./internal/swagger/swagger_pp.json

swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_pub.json ./internal/swagger/swagger_pub.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_pub.json ./internal/swagger/swagger_pub.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_pub.json ./internal/swagger/swagger_pub.json
swagger flatten --with-flatten=remove-unused -q --output=./internal/swagger/swagger_pub.json ./internal/swagger/swagger_pub.json

# read -p "Press enter to continue"