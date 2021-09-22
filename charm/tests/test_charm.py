# Copyright 2021 Canonical Ltd
# See LICENSE file for licensing details.
#
# Learn more about testing at: https://juju.is/docs/sdk/testing

from http.server import BaseHTTPRequestHandler, HTTPServer
import io
import ipaddress
import json
import os
import pathlib
import shutil
import socket
import tarfile
import tempfile
from threading import Thread
import unittest
from unittest.mock import MagicMock, Mock, call, patch

from charm import JimmCharm
import hvac
from ops.model import ActiveStatus, BlockedStatus, MaintenanceStatus, WaitingStatus
from ops.testing import Harness


class TestCharm(unittest.TestCase):
    def setUp(self):
        self.harness = Harness(JimmCharm)
        self.addCleanup(self.harness.cleanup)
        self.harness.begin()
        self.harness.charm._snap = Mock()
        self.harness.charm._systemctl = Mock()
        self.chownmock = patch("os.chown").start()
        self.tempdir = tempfile.TemporaryDirectory()
        self.addCleanup(self.tempdir.cleanup)
        self.harness.charm._dashboard_path = self.tempdir.name + "/dashboard"
        shutil.copytree(os.path.join(self.harness.charm.charm_dir, "templates"),
                        os.path.join(self.tempdir.name, "templates"))
        self.harness.charm.framework.charm_dir = pathlib.Path(self.tempdir.name)

    def dashboard_tarfile(self):
        dashboard_archive = io.BytesIO()

        data = bytes('Hello world', 'utf-8')
        f = io.BytesIO(initial_bytes=data)
        with tarfile.open(fileobj=dashboard_archive, mode='w:bz2') as tar:
            info = tarfile.TarInfo('README.md')
            info.size = len(data)
            tar.addfile(info, f)
            tar.close()

        dashboard_archive.flush()
        dashboard_archive.seek(0)
        data = dashboard_archive.read()
        return data

    def test_install(self):
        service_file = os.path.join(self.harness.charm.charm_dir, 'jimm.service')
        self.harness.add_resource("jimm-snap", "Test data")
        self.harness.add_resource("dashboard", self.dashboard_tarfile())
        self.harness.charm.on.install.emit()
        self.assertTrue(os.path.exists(service_file))
        self.assertEqual(self.harness.charm._snap.call_args.args[0], "install")
        self.assertEqual(self.harness.charm._snap.call_args.args[1], "--dangerous")
        self.assertTrue(str(self.harness.charm._snap.call_args.args[2]).endswith("jimm.snap"))
        self.chownmock.assert_has_calls([
            call(self.tempdir.name + '/dashboard.new/README.md', 0, 0)
        ])

    def test_start(self):
        self.harness.charm.on.start.emit()
        self.harness.charm._systemctl.assert_called_once_with(
            "enable", str(self.harness.charm.service_file))

    def test_start_ready(self):
        with open(self.harness.charm._env_filename(), "wt") as f:
            f.write("test")
        with open(self.harness.charm._env_filename("db"), "wt") as f:
            f.write("test")
        self.harness.charm.on.start.emit()
        self.harness.charm._systemctl.assert_has_calls((
            call("enable", str(self.harness.charm.service_file)),
            call("is-enabled", self.harness.charm.service),
            call("start", self.harness.charm.service)
        ))

    def test_upgrade_charm(self):
        service_file = os.path.join(self.harness.charm.charm_dir, 'jimm.service')
        self.harness.add_resource("jimm-snap", "Test data")
        self.harness.add_resource("dashboard", self.dashboard_tarfile())
        self.harness.charm.on.upgrade_charm.emit()
        self.assertTrue(os.path.exists(service_file))
        self.assertEqual(self.harness.charm._snap.call_args.args[0], 'install')
        self.assertEqual(self.harness.charm._snap.call_args.args[1], '--dangerous')
        self.assertTrue(str(self.harness.charm._snap.call_args.args[2]).endswith("jimm.snap"))
        self.chownmock.assert_has_calls([
            call(self.tempdir.name + '/dashboard.new/README.md', 0, 0)
        ])

    def test_upgrade_charm_ready(self):
        service_file = os.path.join(self.harness.charm.charm_dir, 'jimm.service')
        self.harness.add_resource("jimm-snap", "Test data")
        with open(self.harness.charm._env_filename(), "wt") as f:
            f.write("test")
        with open(self.harness.charm._env_filename("db"), "wt") as f:
            f.write("test")
        self.harness.charm.on.upgrade_charm.emit()
        self.assertTrue(os.path.exists(service_file))
        self.assertEqual(self.harness.charm._snap.call_args.args[0], 'install')
        self.assertEqual(self.harness.charm._snap.call_args.args[1], '--dangerous')
        self.assertTrue(str(self.harness.charm._snap.call_args.args[2]).endswith("jimm.snap"))
        self.harness.charm._systemctl.assert_has_calls((
            call('is-enabled', self.harness.charm.service),
            call('restart', self.harness.charm.service)
        ))

    def test_config_changed(self):
        config_file = os.path.join(self.harness.charm.charm_dir, 'jimm.env')
        os.mkdir(self.tempdir.name + "/dashboard")
        self.harness.update_config({
            "candid-url": "https://candid.example.com",
            "controller-admins": "user1 user2 group1",
            "uuid": "caaa4ba4-e2b5-40dd-9bf3-2bd26d6e17aa",
        })
        self.assertTrue(os.path.exists(config_file))
        with open(config_file) as f:
            lines = f.readlines()
        os.unlink(config_file)
        self.assertEqual(len(lines), 6)
        self.assertEqual(lines[0].strip(), "BAKERY_AGENT_FILE=")
        self.assertEqual(lines[1].strip(), "CANDID_URL=https://candid.example.com")
        self.assertEqual(lines[2].strip(), "JIMM_ADMINS=user1 user2 group1")
        self.assertEqual(lines[3].strip(), "JIMM_UUID=caaa4ba4-e2b5-40dd-9bf3-2bd26d6e17aa")
        self.assertEqual(lines[5].strip(),
                         "JIMM_DASHBOARD_LOCATION=" + self.tempdir.name + "/dashboard")

    def test_config_changed_ready(self):
        config_file = os.path.join(self.harness.charm.charm_dir, 'jimm.env')
        os.mkdir(self.tempdir.name + "/dashboard")
        with open(self.harness.charm._env_filename("db"), "wt") as f:
            f.write("test")
        self.harness.update_config({
            "candid-url": "https://candid.example.com",
            "controller-admins": "user1 user2 group1",
            "uuid": "caaa4ba4-e2b5-40dd-9bf3-2bd26d6e17aa",
        })
        self.assertTrue(os.path.exists(config_file))
        with open(config_file) as f:
            lines = f.readlines()
        os.unlink(config_file)
        self.assertEqual(len(lines), 6)
        self.assertEqual(lines[0].strip(), "BAKERY_AGENT_FILE=")
        self.assertEqual(lines[1].strip(), "CANDID_URL=https://candid.example.com")
        self.assertEqual(lines[2].strip(), "JIMM_ADMINS=user1 user2 group1")
        self.assertEqual(lines[3].strip(), "JIMM_UUID=caaa4ba4-e2b5-40dd-9bf3-2bd26d6e17aa")
        self.assertEqual(lines[5].strip(),
                         "JIMM_DASHBOARD_LOCATION=" + self.tempdir.name + "/dashboard")

    def test_config_changed_with_agent(self):
        config_file = os.path.join(self.harness.charm.charm_dir, 'jimm.env')
        self.harness.charm._agent_filename = os.path.join(self.tempdir.name, "agent.json")
        self.harness.update_config({
            "candid-agent-username": "username@candid",
            "candid-agent-private-key": "agent-private-key",
            "candid-agent-public-key": "agent-public-key",
            "candid-url": "https://candid.example.com",
            "controller-admins": "user1 user2 group1",
            "uuid": "caaa4ba4-e2b5-40dd-9bf3-2bd26d6e17aa",
        })
        self.assertTrue(os.path.exists(self.harness.charm._agent_filename))
        with open(self.harness.charm._agent_filename) as f:
            data = json.load(f)
        self.assertEqual(data["key"]["public"], "agent-public-key")
        self.assertEqual(data["key"]["private"], "agent-private-key")

        self.assertTrue(os.path.exists(config_file))
        with open(config_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 4)
        self.assertEqual(lines[0].strip(),
                         "BAKERY_AGENT_FILE=" + self.harness.charm._agent_filename)
        self.assertEqual(lines[1].strip(), "CANDID_URL=https://candid.example.com")
        self.assertEqual(lines[2].strip(), "JIMM_ADMINS=user1 user2 group1")
        self.assertEqual(lines[3].strip(), "JIMM_UUID=caaa4ba4-e2b5-40dd-9bf3-2bd26d6e17aa")
        self.harness.charm._agent_filename = \
            os.path.join(self.tempdir.name, "no-such-dir", "agent.json")
        self.harness.update_config({
            "candid-agent-username": "username@candid",
            "candid-agent-private-key": "agent-private-key2",
            "candid-agent-public-key": "agent-public-key2",
            "candid-url": "https://candid.example.com",
            "controller-admins": "user1 user2 group1",
            "uuid": "caaa4ba4-e2b5-40dd-9bf3-2bd26d6e17aa",
        })
        with open(config_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 4)
        self.assertEqual(lines[0].strip(), "BAKERY_AGENT_FILE=")
        self.assertEqual(lines[1].strip(), "CANDID_URL=https://candid.example.com")
        self.assertEqual(lines[2].strip(), "JIMM_ADMINS=user1 user2 group1")
        self.assertEqual(lines[3].strip(), "JIMM_UUID=caaa4ba4-e2b5-40dd-9bf3-2bd26d6e17aa")

    def test_leader_elected(self):
        leader_file = os.path.join(self.harness.charm.charm_dir, 'jimm-leader.env')
        self.harness.charm.on.leader_elected.emit()
        with open(leader_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(lines[0].strip(), "JIMM_WATCH_CONTROLLERS=")
        self.harness.set_leader(True)
        with open(leader_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(lines[0].strip(), "JIMM_WATCH_CONTROLLERS=1")

    def test_leader_elected_ready(self):
        leader_file = os.path.join(self.harness.charm.charm_dir, 'jimm-leader.env')
        with open(self.harness.charm._env_filename(), "wt") as f:
            f.write("test")
        with open(self.harness.charm._env_filename("db"), "wt") as f:
            f.write("test")
        self.harness.charm.on.leader_elected.emit()
        with open(leader_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(lines[0].strip(), "JIMM_WATCH_CONTROLLERS=")
        self.harness.set_leader(True)
        with open(leader_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(lines[0].strip(), "JIMM_WATCH_CONTROLLERS=1")
        self.harness.charm._systemctl.assert_has_calls((
            call('is-enabled', self.harness.charm.service),
            call('restart', self.harness.charm.service)
        ))

    def test_db_relation_changed(self):
        db_file = os.path.join(self.harness.charm.charm_dir, 'jimm-db.env')
        id = self.harness.add_relation('db', 'postgresql')
        self.harness.add_relation_unit(id, 'postgresql/0')
        self.harness.update_relation_data(id, 'postgresql/0',
                                          {"master": "host=localhost port=5432"})
        with open(db_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(lines[0].strip(), "JIMM_DSN=pgx:host=localhost port=5432")
        self.harness.update_relation_data(id, 'postgresql/0',
                                          {"master": ""})
        with open(db_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(lines[0].strip(), "JIMM_DSN=pgx:host=localhost port=5432")

    def test_db_relation_changed_ready(self):
        db_file = os.path.join(self.harness.charm.charm_dir, 'jimm-db.env')
        with open(self.harness.charm._env_filename(), "wt") as f:
            f.write("test")
        id = self.harness.add_relation('db', 'postgresql')
        self.harness.add_relation_unit(id, 'postgresql/0')
        self.harness.update_relation_data(id, 'postgresql/0',
                                          {"master": "host=localhost port=5432"})
        with open(db_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(lines[0].strip(), "JIMM_DSN=pgx:host=localhost port=5432")
        self.harness.update_relation_data(id, 'postgresql/0',
                                          {"master": ""})
        with open(db_file) as f:
            lines = f.readlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(lines[0].strip(), "JIMM_DSN=pgx:host=localhost port=5432")
        self.harness.charm._systemctl.assert_has_calls((
            call('is-enabled', self.harness.charm.service),
            call('restart', self.harness.charm.service)
        ))

    def test_website_relation_joined(self):
        id = self.harness.add_relation('website', 'apache2')
        self.harness.add_relation_unit(id, 'apache2/0')
        data = self.harness.get_relation_data(id, self.harness.charm.unit.name)
        self.assertTrue(data)
        self.assertEqual(data["port"], "8080")

    def test_vault_relation_joined(self):
        self.harness.model.get_binding = MagicMock()
        self.harness.model.get_binding().network.egress_subnets[0].network_address = \
            ipaddress.IPv4Address('127.0.0.253')
        id = self.harness.add_relation('vault', 'vault')
        self.harness.add_relation_unit(id, 'vault/0')
        data = self.harness.get_relation_data(id, self.harness.charm.unit.name)
        self.assertTrue(data)
        self.assertEqual(data["secret_backend"], '"charm-jimm-creds"')
        self.assertEqual(data["hostname"], '"{}"'.format(socket.gethostname()))
        self.assertEqual(data["access_address"], '"127.0.0.253"')
        self.assertEqual(data["isolated"], "false")

    def test_vault_relation_changed(self):
        self.harness.charm._vault_secret_filename = os.path.join(self.tempdir.name, 'vault.json')
        self.harness.model.get_binding = MagicMock()
        self.harness.model.get_binding().network.egress_subnets[0].network_address = \
            ipaddress.IPv4Address('127.0.0.253')
        id = self.harness.add_relation('vault', 'vault')
        self.harness.add_relation_unit(id, 'vault/0')
        data = self.harness.get_relation_data(id, self.harness.charm.unit.name)
        self.assertTrue(data)
        hvac.Client = Mock()
        hvac.Client(url="http://vault:8200", token="test-token").sys.unwrap = \
            Mock(return_value={"data": {"secret_id": "test-secret"}})
        self.harness.update_relation_data(id, 'vault/0', {
            "vault_url": '"http://vault:8200"',
            "{}_role_id".format(self.harness.model.unit.name): '"test-role-id"',
            "{}_token".format(self.harness.model.unit.name): '"test-token"'})
        with open(self.harness.charm._vault_secret_filename) as f:
            data = json.load(f)
        self.assertEqual(data, {"data": {"role_id": "test-role-id", "secret_id": "test-secret"}})
        with open(self.harness.charm._env_filename("vault")) as f:
            lines = f.readlines()
        self.assertEqual(lines[0].strip(), "VAULT_ADDR=http://vault:8200")
        self.assertEqual(lines[1].strip(), "VAULT_PATH=charm-jimm-creds")
        self.assertEqual(
            lines[2].strip(),
            "VAULT_SECRET_FILE={}".format(self.harness.charm._vault_secret_filename))
        self.assertEqual(lines[3].strip(), "VAULT_AUTH_PATH=/auth/approle/login")

    def test_stop(self):
        self.harness.charm.on.stop.emit()
        self.harness.charm._systemctl.assert_has_calls((
            call('is-enabled', self.harness.charm.service),
            call('stop', self.harness.charm.service),
            call('is-enabled', self.harness.charm.service),
            call('disable', self.harness.charm.service)
        ))

    def test_update_status(self):
        self.harness.charm._workload_filename = os.path.join(self.tempdir.name, 'jimm.bin')
        self.harness.charm.on.update_status.emit()
        self.assertEqual(self.harness.charm.unit.status,
                         BlockedStatus("waiting for jimm-snap resource"))
        with open(self.harness.charm._workload_filename, "wt") as f:
            f.write("jimm.bin")
        self.harness.charm.on.update_status.emit()
        self.assertEqual(self.harness.charm.unit.status,
                         BlockedStatus("waiting for database"))
        id = self.harness.add_relation('db', 'postgresql')
        self.harness.add_relation_unit(id, 'postgresql/0')
        self.harness.charm.on.update_status.emit()
        self.assertEqual(self.harness.charm.unit.status,
                         WaitingStatus("waiting for database"))
        self.harness.update_relation_data(id, 'postgresql/0',
                                          {"master": "host=localhost port=5432"})
        self.harness.charm.on.update_status.emit()
        self.assertEqual(self.harness.charm.unit.status,
                         MaintenanceStatus("starting"))
        s = HTTPServer(("", 8080), VersionHTTPRequestHandler)
        t = Thread(target=s.serve_forever)
        t.start()
        self.harness.charm.on.update_status.emit()
        s.shutdown()
        s.server_close()
        t.join()
        self.assertEqual(self.harness.charm.unit.status,
                         ActiveStatus())


class VersionHTTPRequestHandler(BaseHTTPRequestHandler):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        s = json.dumps({"Version": "1.2.3"})
        self.wfile.write(s.encode("utf-8"))
