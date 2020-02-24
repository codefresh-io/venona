const _ = require('lodash');
const Job = require('./../BaseJob');
const { create: createLogger } = require('../../services/Logger');

jest.mock('./../../services/Logger');

describe('BaseJob unit tests', () => {

	it('Should construct', () => {
		const job = new Job(_.noop(), _.noop(), createLogger());
		expect(Object.keys(job).sort()).toEqual(['codefreshAPI', 'runtimes', 'logger'].sort());
	});

	it('Should return the requested runtime', async () => {
		const kubernetesAPI = {};
		const runtimes = {
			'runtime': {
				name: 'runtime',
				kubernetesAPI,
			},
		};
		const job = new Job(_.noop(), runtimes, createLogger());
		await expect(job.getKubernetesService('runtime')).resolves.toEqual(kubernetesAPI);
	});

	it('Should throw an error in case the runtime was not found', async () => {
		const runtimes = {};
		const job = new Job(_.noop(), runtimes, createLogger());
		await expect(job.getKubernetesService('runtime')).rejects.toThrowError('Kubernetes client for runtime "runtime" was not found');
	});

	it('Should throw an error in case the runtime\'s KubernetesAPI was not initialized', async () => {
		const runtimes = {
			'runtime': {
				metadata: {
					error: 'failed to initialize runtime'
				}
			}
		};
		const job = new Job(_.noop(), runtimes, createLogger());
		await expect(job.getKubernetesService('runtime')).rejects.toThrowError('Kubernetes client for runtime "runtime" was not created due to error: failed to initialize runtime');
	});
});
