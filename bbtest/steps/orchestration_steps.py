#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import time
from behave import *
from helpers.shell import execute
from helpers.eventually import eventually
from helpers.http import Request


@given('package {package} is {operation}')
def step_impl(context, package, operation):
  if operation == 'installed':
    (code, result, error) = execute(["apt-get", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", "-o=Dpkg::Options::=--force-confold", context.unit.binary])
    assert code == 'OK', "unable to install with code {} and {} {}".format(code, result, error)
    assert os.path.isfile('/etc/bondster-bco/conf.d/init.conf') is True, 'config file does not exists'
    execute(['systemctl', 'start', package])
  elif operation == 'uninstalled':
    execute(['systemctl', 'stop', package])
    (code, result, error) = execute(["apt-get", "-f", "-qq", "remove", package])
    assert code == 'OK', "unable to uninstall with code {} and {} {}".format(code, result, error)
    (code, result, error) = execute(["apt-get", "-f", "-qq", "purge", package])
    assert code == 'OK', "unable to purge with code {} and {} {}".format(code, result, error)
    assert os.path.isfile('/etc/bondster-bco/conf.d/init.conf') is False, 'config file still exists'
  else:
    assert False, 'unknown operation {}'.format(operation)


@given('systemctl contains following active units')
@then('systemctl contains following active units')
def step_impl(context):
  (code, result, error) = execute(["systemctl", "list-units", "--all", "--no-legend", "--state=active"])
  assert code == 'OK', str(result) + ' ' + str(error)
  items = []
  for row in context.table:
    items.append(row['name'] + '.' + row['type'])
  result = [item.replace('*', '').strip().split(' ')[0].strip() for item in result.split(os.linesep)]
  result = [item for item in result if item in items]
  assert len(result) > 0, 'units not found\n{}'.format(result)


@given('systemctl does not contain following active units')
@then('systemctl does not contain following active units')
def step_impl(context):
  (code, result, error) = execute(["systemctl", "list-units", "--all", "--no-legend", "--state=active"])
  assert code == 'OK', str(result) + ' ' + str(error)
  items = []
  for row in context.table:
    items.append(row['name'] + '.' + row['type'])
  result = [item.replace('*', '').strip().split(' ')[0].strip() for item in result.split(os.linesep)]
  result = [item for item in result if item in items]
  assert len(result) == 0, 'units found\n{}'.format(result)


@given('unit "{unit}" is running')
@then('unit "{unit}" is running')
def unit_running(context, unit):
  @eventually(10)
  def wait_for_unit_state_change():
    (code, result, error) = execute(["systemctl", "show", "-p", "SubState", unit])
    assert code == 'OK', str(result) + ' ' + str(error)
    assert 'SubState=running' in result, result

  wait_for_unit_state_change()

  if 'bondster-bco-rest' in unit:
    request = Request(method='GET', url="https://127.0.0.1/health")

    @eventually(5)
    def wait_for_healthy():
      response = request.do()
      assert response.status == 200, str(response.status)

    wait_for_healthy()


@given('unit "{unit}" is not running')
@then('unit "{unit}" is not running')
def unit_not_running(context, unit):
  @eventually(10)
  def wait_for_unit_state_change():
    (code, result, error) = execute(["systemctl", "show", "-p", "SubState", unit])
    assert code == 'OK', str(result) + ' ' + str(error)
    assert 'SubState=dead' in result, result

  wait_for_unit_state_change()


@given('{operation} unit "{unit}"')
@when('{operation} unit "{unit}"')
def operation_unit(context, operation, unit):
  (code, result, error) = execute(["systemctl", operation, unit])
  assert code == 'OK', str(result) + ' ' + str(error)


@given('{unit} is configured with')
def unit_is_configured(context, unit):
  params = dict()
  for row in context.table:
    params[row['property']] = row['value']
  context.unit.configure(params)


@then('tenant {tenant} is offboarded')
@given('tenant {tenant} is offboarded')
def offboard_unit(context, tenant):
  logfile = os.path.realpath('{}/../../reports/blackbox-tests/logs/bondster-bco-import.{}.log'.format(os.path.dirname(__file__), tenant))
  (code, result, error) = execute(['journalctl', '-o', 'cat', '-u', 'bondster-bco-import@{}.service'.format(tenant), '--no-pager'])
  if code == 'OK' and result:
    with open(logfile, 'w') as f:
      f.write(result)
  execute(['systemctl', 'stop', 'bondster-bco-import@{}.service'.format(tenant)])
  (code, result, error) = execute(['journalctl', '-o', 'cat', '-u', 'bondster-bco-import@{}.service'.format(tenant), '--no-pager'])
  if code == 'OK' and result:
    with open(logfile, 'w') as fd:
      fd.write(result)
  execute(['systemctl', 'disable', 'bondster-bco-import@{}.service'.format(tenant)])
  unit_not_running(context, 'bondster-bco-import@{}'.format(tenant))


@given('tenant {tenant} is onboarded')
def onboard_unit(context, tenant):
  (code, result, error) = execute(["systemctl", 'enable', 'bondster-bco-import@{}'.format(tenant)])
  assert code == 'OK', str(result) + ' ' + str(error)
  (code, result, error) = execute(["systemctl", 'start', 'bondster-bco-import@{}'.format(tenant)])
  assert code == 'OK', str(result) + ' ' + str(error)
  unit_running(context, 'bondster-bco-import@{}'.format(tenant))
