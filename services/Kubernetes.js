const { Client, config } = require('kubernetes-client');
const utils = require('./../utils');

const ERROR_MESSAGES = {
	MISSING_KUBERNETES_URL: 'Failed to construct Kubernetes API service, missing Kubernetes URL',
	MISSING_KUBERNETES_BEARER_TOKEN: 'Failed to construct Kubernetes API service, missing Kubernetes bearer token',
	MISSING_KUBERNETES_CA_CERTIFICATE: 'Failed to construct Kubernetes API service, missing Kubernetes ca certificate',
	FAILED_TO_INIT: 'Failed to complete Kubernetes service initialization',
	FAILED_TO_CREATE_POD: 'Failed to create Kubernetes pod',
	FAILED_TO_DELETE_POD: 'Failed to delete Kubernetes pod',
};


class Kubernetes {
	constructor(metadata, client) {
		this.metadata = metadata;
		this.client = client;
	}

	static buildFromInCluster(metadata) {
		const client = new Client({ config: config.getInCluster() });
		return new this(metadata, client);
	}

	static buildFromConfig(metadata, options) {
		const url = utils.getPropertyOrError(options, 'config.url', ERROR_MESSAGES.MISSING_KUBERNETES_URL);
		const bearer = utils.getPropertyOrError(options, 'config.auth.bearer', ERROR_MESSAGES.MISSING_KUBERNETES_BEARER_TOKEN);
		const ca = utils.getPropertyOrError(options, 'config.ca', ERROR_MESSAGES.MISSING_KUBERNETES_CA_CERTIFICATE);
		const client = new Client({
			config: {
				url,
				auth: {
					bearer: Buffer.from(bearer, 'base64'),
				},
				ca: Buffer.from(ca, 'base64'),
			},
		});
		return new this(metadata, client);
	}

	async init() {
		try {
			await this.client.loadSpec();
			return Promise.resolve();
		} catch (err) {
			throw new Error(`${ERROR_MESSAGES.FAILED_TO_INIT} with error: ${err.message}`);
		}
	}

	async createPod(logger, spec) {
		try {
			await this.client.api.v1.namespaces(spec.metadata.namespace).pod.post({ body: spec });
			logger.info('Pod created');
			return Promise.resolve();
		} catch (err) {
			throw new Error(`${ERROR_MESSAGES.FAILED_TO_CREATE_POD} with message: ${err.message}`);
		}
	}

	async deletePod(logger, namespace, name) {
		try {
			await this.client.api.v1.namespaces(namespace).pod(name).delete();
			logger.info('Pod deleted');
			return Promise.resolve();
		} catch (err) {
			throw new Error(`${ERROR_MESSAGES.FAILED_TO_DELETE_POD} with message: ${err.message}`);
		}
	}
}

Kubernetes.Errors = ERROR_MESSAGES;

module.exports = Kubernetes;
