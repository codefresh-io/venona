const _ = require('lodash');
const { create: createLogger } = require('../../../../services/Logger');
const StartWorkflow = require('../startWorkflow');

jest.mock('./../../../../services/Logger');

describe('StartWorkflow unit tests', () => {
	it('Should throw an error when call to Kubernetes service is been rejected', () => {
		const logger = createLogger();
		const kubernetesAPIMock = {
			createPod: jest.fn().mockRejectedValue(new Error('Error!!!')),
		};
		const taskDef = {};
		const task = new StartWorkflow(_.noop(), kubernetesAPIMock, logger);
		return expect(task.run(taskDef)).rejects.toThrowError('Failed to run task StartWorkflow, failed to create pod');
	});

	it('Should call twice to Kubernetes service', () => {
		const logger = createLogger();
		const spy = jest.fn().mockResolvedValue();
		const kubernetesAPIMock = {
			createPod: spy,
		};
		const taskDef = {
			runtime: {},
			dockerDaemon: {},
		};
		const task = new StartWorkflow(_.noop(), kubernetesAPIMock, logger);
		return task.run(taskDef)
			.then(() => {
				const loggerMacher = expect.objectContaining({
					error: expect.any(Function),
					info: expect.any(Function),
					child: expect.any(Function),
				});
				expect(spy).toHaveBeenCalledTimes(2);
				expect(spy).toHaveBeenNthCalledWith(1, loggerMacher, taskDef.runtime);
				expect(spy).toHaveBeenNthCalledWith(2, loggerMacher, taskDef.dockerDaemon);
			});
	});

	it('Should return value', () => {
		const logger = createLogger();
		const spy = jest.fn().mockResolvedValue({});
		const kubernetesAPIMock = {
			createPod: spy,
		};
		const taskDef = {
			runtime: {},
			dockerDaemon: {},
		};
		const task = new StartWorkflow(_.noop(), kubernetesAPIMock, logger);
		return expect(task.run(taskDef)).resolves.toEqual({ dind: {}, runtime: {}});
	});
});
