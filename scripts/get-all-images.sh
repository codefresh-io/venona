MYDIR=$(dirname $0)
CHARTDIR="${MYDIR}/../charts/cf-runtime"
VALUESFILE="../charts/cf-runtime/.ci/values-ci.yaml"
OUTPUTFILE=$1
helm dependency update $CHARTDIR
helm template --values $VALUESFILE --set global.runtimeName="dummy" $CHARTDIR | grep -E 'image: | dindImage:' | awk -F ': ' '{print $2}' | tr -d '"' | uniq > $OUTPUTFILE