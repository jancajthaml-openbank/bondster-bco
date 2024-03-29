#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from helpers.unit import UnitHelper
from helpers.zmq import ZMQHelper
from mocks.bondster.server import BondsterMock
from mocks.vault.server import VaultMock
from mocks.ledger.server import LedgerMock
from openbank_testkit import StatsdMock
from helpers.logger import logger


def before_feature(context, feature):
  context.statsd.clear()
  context.log.info('')
  context.log.info('  (FEATURE) {}'.format(feature.name))


def before_scenario(context, scenario):
  context.log.info('')
  context.log.info('  (SCENARIO) {}'.format(scenario.name))
  context.log.info('')


def after_scenario(context, feature):
  context.unit.collect_logs()


def after_feature(context, feature):
  context.zmq.clear()


def before_all(context):
  context.log = logger()
  context.log.info('')
  context.log.info('  (START)')
  context.tokens = dict()
  context.unit = UnitHelper(context)
  context.zmq = ZMQHelper(context)
  context.bondster = BondsterMock(context)
  context.ledger = LedgerMock(context)
  context.vault = VaultMock(context)
  context.statsd = StatsdMock()
  context.statsd.start()
  context.bondster.start()
  context.ledger.start()
  context.vault.start()
  context.zmq.start()
  context.unit.configure()
  context.unit.download()


def after_all(context):
  context.log.info('')
  context.log.info('  (END)')
  context.log.info('')
  context.bondster.stop()
  context.ledger.stop()
  context.vault.stop()
  context.unit.teardown()
  context.zmq.stop()
  context.statsd.stop()
