const _ = require('lodash');
const { create: createLogger } = require('../../../../services/Logger');
const DeletePodTask = require('../DeletePod.task');

jest.mock('./../../../../services/Logger');

const getValidTaskDef = () => {
	return {
		spec: {
			namespace: 'namespace',
			name: 'docker-daemon-name'
		}
	};
};

describe('DeletePod task unit tests', () => {

	describe('negative', () => {

		it('Should throw an error when call to Kubernetes service is been rejected', () => {
			const logger = createLogger();
			const kubernetesAPIMock = {
				deletePod: jest.fn().mockRejectedValue(new Error('Error!!!')),
			};

			const task = new DeletePodTask(_.noop(), kubernetesAPIMock, logger);
			return expect(task.exec(getValidTaskDef())).rejects.toThrowError('Failed to run task DeletePod: Error!!!');
		});

		describe('validation', () => {
			it('should throw an error in case spec is missing', () => {
				const logger = createLogger();
				const spy = jest.fn().mockResolvedValue();
				const kubernetesAPIMock = {
					deletePod: spy,
				};

				const taskDef = getValidTaskDef();
				delete taskDef.spec;
				const task = new DeletePodTask(_.noop(), kubernetesAPIMock, logger);
				return expect(task.exec(taskDef)).rejects.toThrowError('child "spec" fails because ["spec" is required]');
			});

			it('should throw an error in case namespace is missing', () => {
				const logger = createLogger();
				const spy = jest.fn().mockResolvedValue();
				const kubernetesAPIMock = {
					deletePod: spy,
				};

				const taskDef = getValidTaskDef();
				delete taskDef.spec.namespace;
				const task = new DeletePodTask(_.noop(), kubernetesAPIMock, logger);
				return expect(task.exec(taskDef)).rejects.toThrowError('child "spec" fails because [child "namespace" fails because ["namespace" is required]]');
			});

			it('should throw an error in case name is missing', () => {
				const logger = createLogger();
				const spy = jest.fn().mockResolvedValue();
				const kubernetesAPIMock = {
					deletePod: spy,
				};

				const taskDef = getValidTaskDef();
				delete taskDef.spec.name;
				const task = new DeletePodTask(_.noop(), kubernetesAPIMock, logger);
				return expect(task.exec(taskDef)).rejects.toThrowError('child "spec" fails because [child "name" fails because ["name" is required]]');
			});
		});
	});

	describe('positive', () => {
		it('Should call twice to Kubernetes service', () => {
			const logger = createLogger();
			const spy = jest.fn().mockResolvedValue();
			const kubernetesAPIMock = {
				deletePod: spy,
			};

			const taskDef = getValidTaskDef();
			const task = new DeletePodTask(_.noop(), kubernetesAPIMock, logger);
			return task.exec(taskDef)
				.then(() => {
					const loggerMacher = expect.objectContaining({
						error: expect.any(Function),
						info: expect.any(Function),
						child: expect.any(Function),
					});
					expect(spy).toHaveBeenCalledTimes(1);
					expect(spy).toHaveBeenNthCalledWith(1, loggerMacher, taskDef.spec.namespace, taskDef.spec.name);
				});
		});

		it('Should return value', () => {
			const logger = createLogger();
			const spy = jest.fn().mockResolvedValue({});
			const kubernetesAPIMock = {
				deletePod: spy,
			};
			const taskDef = getValidTaskDef();
			const task = new DeletePodTask(_.noop(), kubernetesAPIMock, logger);
			return expect(task.exec(taskDef)).resolves.toEqual('OK');
		});

		it('should succeed in case of getting a 404 error when trying to delete the pod', () => {
			const logger = createLogger();
			const error = new Error('Error!!!');
			error.code = 404;
			const kubernetesAPIMock = {
				deletePod: jest.fn().mockRejectedValue(error),
			};

			const taskDef = getValidTaskDef();
			const task = new DeletePodTask(_.noop(), kubernetesAPIMock, logger);
			return expect(task.exec(taskDef)).resolves.toEqual('OK');
		});
	});
});
