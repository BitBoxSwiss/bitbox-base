docker build --tag=digitalbitbox/bitbox-base .
docker run -v ${TRAVIS_BUILD_DIR}:/opt/go/src/github.com/digitalbitbox/bitbox-base/ \
        -i digitalbitbox/bitbox-base \
        bash -c "make -C \$GOPATH/src/github.com/digitalbitbox/bitbox-base/middleware ci" \
        bash -c "make -C \$GOPATH/src/github.com/digitalbitbox/bitbox-base/middleware native" \
        bash -c "make -C \$GOPATH/src/github.com/digitalbitbox/bitbox-base/tools/bbbfancontrol"

