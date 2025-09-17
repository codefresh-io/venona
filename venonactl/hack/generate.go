package hack

/*
We are using generated template.go for serialized kubernetes assets
*/
//go:generate go run github.com/codefresh-io/venona/venonactl/pkg/templates kubernetes
//go:generate go run github.com/codefresh-io/venona/venonactl/pkg/obj kubernetes
