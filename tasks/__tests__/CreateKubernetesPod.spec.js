const _ = require('lodash');
const { create: createLogger } = require('../../services/Logger');
const CreateKubernetesPod = require('./../CreateKubernetesPod');

jest.mock('./../../services/Logger');

describe('CreateKubernetesPod unit tests', () => {
	it('Should call kubernetes service', () => {
		const kubernetesAPI = {
			createPod: jest.fn().mockResolvedValue(),
		};
		const podDef = 'pod-def';
		const logger = createLogger();
		const task = new CreateKubernetesPod(_.noop(), kubernetesAPI, logger);
		return task.run(podDef).then(() => {
			expect(kubernetesAPI.createPod).toHaveBeenCalled();
			expect(kubernetesAPI.createPod).toHaveBeenCalledWith(expect.objectContaining({
				error: expect.any(Function),
				info: expect.any(Function),
				child: expect.any(Function),
			}), podDef);
		});
	});

	it('Should sent to logger propper message', () => {
		const kubernetesAPI = {
			createPod: jest.fn().mockResolvedValue(),
		};
		const podDef = 'pod-def';
		const logger = createLogger();
		const task = new CreateKubernetesPod(_.noop(), kubernetesAPI, logger);
		return task.run({ spec: podDef }).then(() => {
			expect(logger.child.mock.results[1].value.info.mock.calls[0][0]).toEqual('Running CreateKubernetesPod task');
		});
	});

	it('Should throw an error when call to kuberentes service been rejected', () => {
		const kubernetesAPI = {
			createPod: jest.fn().mockRejectedValue(new Error('Error!')),
		};
		const podDef = 'pod-def';
		const logger = createLogger();
		const task = new CreateKubernetesPod(_.noop(), kubernetesAPI, logger);
		return expect(task.run({ spec: podDef })).rejects.toThrowError('Failed to run task CreateKubernetesPod, failed to create pod with message:');
	});

	it('Should log a message error when call to kuberentes service been rejected', () => {
		const kubernetesAPI = {
			createPod: jest.fn().mockRejectedValue(new Error('Error!')),
		};
		const podDef = 'pod-def';
		const logger = createLogger();
		const task = new CreateKubernetesPod(_.noop(), kubernetesAPI, logger);
		return task.run({ spec: podDef })
			.catch(() => {
				expect(logger.child.mock.results[1].value.error.mock.calls[0][0]).toMatch('Failed to run task CreateKubernetesPod, failed to create pod with message:');
			});
	});
});
