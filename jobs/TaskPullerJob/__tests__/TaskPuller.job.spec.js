const _ = require('lodash');
const Promise = require('bluebird');
const { create: createLogger } = require('../../../services/Logger');
const TaskPullerJob = require('../TaskPuller.job');
const CreatePodTask = require('../tasks/CreatePod.task');
const DeletePodTask = require('../tasks/DeletePod.task');
const CreatePvcTask = require('../tasks/CreatePvc.task');
const DeletePvcTask = require('../tasks/DeletePvc.task');

jest.mock('./../../../services/Logger');
jest.mock('./../tasks/CreatePod.task');
jest.mock('./../tasks/DeletePod.task');
jest.mock('./../tasks/CreatePvc.task');
jest.mock('./../tasks/DeletePvc.task');

describe('TaskPullerJob unit tests', () => {
	it('Should throw an error when codefresh service call failed', () => {
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockRejectedValue(new Error('Failed')),
		}, _.noop(), logger);
		return expect(task.exec()).rejects.toThrowError('Failed to run job TaskPuller, call to Codefresh rejected with message');
	});

	it('Should pass logger to codefresh api service', () => {
		const spy = jest.fn().mockResolvedValue();
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: spy,
		}, _.noop(), logger);
		return task.exec()
			.then(() => {
				expect(spy).toHaveBeenCalledWith(expect.objectContaining({
					error: expect.any(Function),
					info: expect.any(Function),
					child: expect.any(Function),
				}));
			});
	});

	it('Should map all results to tasks by types and execute them', () => {
		CreatePodTask.mockImplementationOnce(() => {
			return {
				exec: jest.fn(async () => {
					return {
						status: 'ok'
					};
				}),
			};
		});
		const tasks = [
			{
				type: 'CreatePod',
			}
		];
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.exec()).resolves.toEqual([{ status: 'ok'}]);
	});

	it('Should not fail when unknown type task arrives', () => {
		const tasks = [
			{
				type: 'Fake-Task-Type',
			},
		];
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.exec()).resolves.toEqual([]);
	});

	it('Should not fail when non-typed task arrives', () => {
		const tasks = [
			{
				prop: 'a'
			},
		];
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.exec()).resolves.toEqual([]);
	});

	it('Should not fail when a task failed', () => {
		const tasks = [
			{
				type: 'DeletePod',
			},
			{
				type: 'CreatePod',
			}
		];
		DeletePodTask.mockImplementationOnce(() => ({
			exec: jest.fn(async () => Promise.reject('error'))	
		}));
		CreatePodTask.mockImplementationOnce(() => ({
			exec: jest.fn(async () => Promise.resolve('value'))	
		}));
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.exec()).resolves.toEqual(['value', undefined]);
	});

	it('Should not fail when task failed', () => {
		const tasks = [
			{
				prop: 'a'
			},
		];
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.exec()).resolves.toEqual([]);
	});

	it('Should always execute and resolve the HIGH priority task first', async () => {
		const tasks = [
			{
				type: 'DeletePod'
			},
			{
				type: 'DeletePvc'
			},
			{
				type: 'CreatePod'				
			},
			{
				type: 'CreatePvc'
			},
		];

		DeletePodTask.mockImplementationOnce(() => ({
			exec: jest.fn(async () => Promise.delay(10,'DeletePod'))	
		}));
		DeletePvcTask.mockImplementationOnce(() => ({
			exec: jest.fn(async () => Promise.delay(10,'DeletePvc'))	
		}));
		CreatePodTask.mockImplementationOnce(() => ({
			exec: jest.fn(async () => Promise.delay(400, 'CreatePod'))	
		}));
		CreatePvcTask.mockImplementationOnce(() => ({
			exec: jest.fn(async () => Promise.delay(400, 'CreatePvc'))	
		}));

		const logger = createLogger();
		const job = new TaskPullerJob({ 
			pullTasks: jest.fn().mockResolvedValueOnce(tasks) 
		}, _.noop, logger);
		
		const results = await job.exec();
		expect(results).toEqual(['CreatePod', 'CreatePvc', 'DeletePod', 'DeletePvc']);
	});
});
