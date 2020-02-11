const { Client, config } = require('kubernetes-client');
const utils = require('./../utils');
const fs = require('fs');
const yaml = require('js-yaml');
const _ = require('lodash');

const ERROR_MESSAGES = {
	MISSING_KUBERNETES_URL: 'Failed to construct Kubernetes API service, missing Kubernetes URL',
	MISSING_KUBERNETES_BEARER_TOKEN: 'Failed to construct Kubernetes API service, missing Kubernetes bearer token',
	MISSING_KUBERNETES_CA_CERTIFICATE: 'Failed to construct Kubernetes API service, missing Kubernetes ca certificate',
	MISSING_NAMESPACE: 'Failed to get Kubernetes namespace',
	MISSING_VENONA_CONF: 'Failed to get venona configuration',

	FAILED_TO_INIT: 'Failed to complete Kubernetes service initialization',
	FAILED_TO_CREATE_POD: 'Failed to create Kubernetes pod',
	FAILED_TO_DELETE_POD: 'Failed to delete Kubernetes pod',
	FAILED_TO_CREATE_PVC: 'Failed to create Kubernetes pvc',
	FAILED_TO_DELETE_PVC: 'Failed to delete Kubernetes pvc',
};

const RUNTIME_SECRET_LOCATION = '/etc/secrets/venonaconf';

class Kubernetes {
	constructor(metadata, agentClient, runtimes = {}) {
		this.metadata = metadata;
		this.agentClient = agentClient;
		this.runtimes = runtimes;
	}

	static parseRuntimesFromVenonaConf(venonaConf, encoding) {
		let buff = new Buffer(venonaConf, encoding);
		return _.get(yaml.safeLoad(buff.toString()), 'Runtimes');
	}
	
	static async buildFromInCluster(metadata) {
		const client = new Client({ config: config.getInCluster() });
		let venonaConf = '';
		const isVenonaConfExist = await new Promise((resolve) => {
			fs.access(RUNTIME_SECRET_LOCATION, (err) => {
				if (err) {
					resolve(false);
				} else {
					resolve(true);
				}
			});
		});
		if (isVenonaConfExist) {
			venonaConf = await new Promise((resolve, reject) => {
				fs.readFile(RUNTIME_SECRET_LOCATION, (err, data) => {
					if (err) {
						reject(err);
					}else {
						resolve(data);
					}
				});
			});
		}
		return new this(metadata, client,  Kubernetes.parseRuntimesFromVenonaConf(venonaConf));
	}

	static buildFromConfig(metadata, options) {
		const url = utils.getPropertyOrError(options, 'config.url', ERROR_MESSAGES.MISSING_KUBERNETES_URL);
		const bearer = utils.getPropertyOrError(options, 'config.auth.bearer', ERROR_MESSAGES.MISSING_KUBERNETES_BEARER_TOKEN);
		const ca = utils.getPropertyOrError(options, 'config.ca', ERROR_MESSAGES.MISSING_KUBERNETES_CA_CERTIFICATE);
		const venonaConf = utils.getPropertyOrError(options, 'metadata.venonaConf', ERROR_MESSAGES.MISSING_VENONA_CONF);
		const client = new Client({
			config: {
				url,
				auth: {
					bearer: Buffer.from(bearer, 'base64'),
				},
				ca: Buffer.from(ca, 'base64'),
			},
		});
		return new this(metadata, client, Kubernetes.parseRuntimesFromVenonaConf(venonaConf, 'base64') );
	}

	async getClient(runtimeName) {
		if (!this.runtimes[runtimeName]) {
			throw new Error(`runtime ${runtimeName} is not found`);
		}
		const runtimeConfig = this.runtimes[runtimeName];
		if (!runtimeConfig.client) {
			runtimeConfig.client = new Client({
				config: {
					url: runtimeConfig.Host,
					auth: {
						bearer: runtimeConfig.Token,
					},
					ca: runtimeConfig.Crt,
				},
			});
			await runtimeConfig.client.loadSpec();
		}
		return runtimeConfig.client;
	}

	async init() {
		try {
			await this.agentClient.loadSpec();
			return Promise.resolve();
		} catch (err) {
			throw new Error(`${ERROR_MESSAGES.FAILED_TO_INIT} with error: ${err.message}`);
		}
	}

	async createPod(logger, spec, runtime) {
		try {
			const client = await this.getClient(runtime);
			await client.api.v1.namespaces(spec.metadata.namespace).pod.post({ body: spec });
			logger.info('Pod created');
			return Promise.resolve();
		} catch (err) {
			throw new Error(`${ERROR_MESSAGES.FAILED_TO_CREATE_POD} with message: ${err.message}`);
		}
	}

	async deletePod(logger, namespace, name, runtime) {
		try {
			const client = await this.getClient(runtime);
			await client.api.v1.namespaces(namespace).pod(name).delete();
			logger.info('Pod deleted');
			return Promise.resolve();
		} catch (err) {
			throw new Error(`${ERROR_MESSAGES.FAILED_TO_DELETE_POD} with message: ${err.message}`);
		}
	}

	async createPvc(logger, spec, runtime) {
		try {
			const client = await this.getClient(runtime);
			await client.api.v1.namespaces(spec.metadata.namespace).persistentvolumeclaim.post({ body: spec });
			logger.info('Pvc created');
			return Promise.resolve();
		} catch (err) {
			throw new Error(`${ERROR_MESSAGES.FAILED_TO_CREATE_PVC} with message: ${err.message}`);
		}
	}

	async deletePvc(logger, namespace, name, runtime) {
		try {
			const client = await this.getClient(runtime);
			await client.api.v1.namespaces(namespace).persistentvolumeclaim(name).delete();
			logger.info('Pvc deleted');
			return Promise.resolve();
		} catch (err) {
			throw new Error(`${ERROR_MESSAGES.FAILED_TO_DELETE_PVC} with message: ${err.message}`);
		}
	}
}

Kubernetes.Errors = ERROR_MESSAGES;

module.exports = Kubernetes;
