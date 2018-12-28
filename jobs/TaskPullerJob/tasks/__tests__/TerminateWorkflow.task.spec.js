const _ = require('lodash');
const { create: createLogger } = require('../../../../services/Logger');
const TerminateWorkflow = require('../TerminateWorkflow.task');

jest.mock('./../../../../services/Logger');

const getValidTaskDef = () => {
	return {
		dockerDaemon: {
			metadata: {
				name: 'docker-daemon-name',
				namespace: 'namespace'
			},
		},
		runtime: {
			metadata: {
				name: 'runtime-name',
				namespace: 'namespace'
			},
		}
	};
};

describe('TerminateWorkflow unit tests', () => {
	it('Should throw an error when call to Kubernetes service is been rejected', () => {
		const logger = createLogger();
		const kubernetesAPIMock = {
			deletePod: jest.fn().mockRejectedValue(new Error('Error!!!')),
		};

		const task = new TerminateWorkflow(_.noop(), kubernetesAPIMock, logger);
		return expect(task.run(getValidTaskDef())).rejects.toThrowError('Failed to run task TerminateWorkflow, failed to delete pod');
	});

	it('Should throw an error when docker daemon name is not passed', () => {});
	it('Should throw an error when docker daemon namespace is not passed', () => {});
	it('Should throw an error when runtime name is not passed', () => {});
	it('Should throw an error when runtime namespace is not passed', () => {});

	it('Should call twice to Kubernetes service', () => {
		const logger = createLogger();
		const spy = jest.fn().mockResolvedValue();
		const kubernetesAPIMock = {
			deletePod: spy,
		};

		const taskDef = getValidTaskDef();
		const task = new TerminateWorkflow(_.noop(), kubernetesAPIMock, logger);
		return task.run(taskDef)
			.then(() => {
				const loggerMacher = expect.objectContaining({
					error: expect.any(Function),
					info: expect.any(Function),
					child: expect.any(Function),
				});
				expect(spy).toHaveBeenCalledTimes(2);
				expect(spy).toHaveBeenNthCalledWith(1, loggerMacher, taskDef.runtime.metadata.namespace, taskDef.runtime.metadata.name);
				expect(spy).toHaveBeenNthCalledWith(2, loggerMacher, taskDef.dockerDaemon.metadata.namespace, taskDef.dockerDaemon.metadata.name);
			});
	});

	it('Should return value', () => {
		const logger = createLogger();
		const spy = jest.fn().mockResolvedValue({});
		const kubernetesAPIMock = {
			deletePod: spy,
		};
		const taskDef = getValidTaskDef();
		const task = new TerminateWorkflow(_.noop(), kubernetesAPIMock, logger);
		return expect(task.run(taskDef)).resolves.toEqual('OK');
	});
});
