require_relative 'placeholders'

step "bondster-bco is restarted" do ||
  ids = %x(systemctl -t service --no-legend | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("bondster-bco-import@")
  }.map { |x| x.chomp(".service") }

  expect(ids).not_to be_empty

  ids.each { |e|
    %x(systemctl restart #{e} 2>&1)
  }

  ids << "bondster-bco-rest"

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end

step "tenant :tenant is offboarded" do |tenant|
  eventually() {
    %x(journalctl -o short-precise -u bondster-bco-import@#{tenant}.service --no-pager > /reports/bondster-bco@#{tenant}.log 2>&1)
    %x(systemctl stop bondster-bco-import@#{tenant} 2>&1)
    %x(systemctl disable bondster-bco-import@#{tenant} 2>&1)
    %x(journalctl -o short-precise -u bondster-bco-import@#{tenant}.service --no-pager > /reports/bondster-bco@#{tenant}.log 2>&1)
  }
end

step "tenant :tenant is onbdoarded" do |tenant|
  params = [
    "BONDSTER_BCO_STORAGE=/data",
    "BONDSTER_BCO_LOG_LEVEL=DEBUG",
    "BONDSTER_BCO_BONDSTER_GATEWAY=https://127.0.0.1:4000",
    "BONDSTER_BCO_SYNC_RATE=1h",
    "BONDSTER_BCO_VAULT_GATEWAY=https://127.0.0.1:4400",
    "BONDSTER_BCO_LEDGER_GATEWAY=https://127.0.0.1:4401",
    "BONDSTER_BCO_METRICS_OUTPUT=/reports/metrics.json",
    "BONDSTER_BCO_LAKE_HOSTNAME=127.0.0.1",
    "BONDSTER_BCO_METRICS_REFRESHRATE=1h",
    "BONDSTER_BCO_HTTP_PORT=443",
    "BONDSTER_BCO_SECRETS=/opt/bondster-bco/secrets",
    "BONDSTER_BCO_ENCRYPTION_KEY=/opt/bondster-bco/secrets/fs_encryption.key"
  ].join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{params}' > /etc/init/bondster-bco.conf)

  %x(systemctl enable bondster-bco-import@#{tenant} 2>&1)
  %x(systemctl start bondster-bco-import@#{tenant} 2>&1)

  ids = %x(systemctl list-units | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("bondster-bco-")
  }.map { |x| x.chomp(".service") }

  ids.each { |e|
    %x(systemctl restart #{e} 2>&1)
  }

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end

step "bondster-bco is reconfigured with" do |configuration|
  params = Hash[configuration.split("\n").map(&:strip).reject(&:empty?).map {|el| el.split '='}]
  defaults = {
    "STORAGE" => "/data",
    "LOG_LEVEL" => "DEBUG",
    "BONDSTER_GATEWAY" => "https://127.0.0.1:4000",
    "SYNC_RATE" => "1h",
    "VAULT_GATEWAY" => "https://127.0.0.1:4400",
    "LEDGER_GATEWAY" => "https://127.0.0.1:4401",
    "METRICS_OUTPUT" => "/reports/metrics.json",
    "LAKE_HOSTNAME" => "127.0.0.1",
    "METRICS_REFRESHRATE" => "1h",
    "HTTP_PORT" => "443",
    "SECRETS" => "/opt/bondster-bco/secrets",
    "ENCRYPTION_KEY" => "/opt/bondster-bco/secrets/fs_encryption.key"
  }

  config = Array[defaults.merge(params).map {|k,v| "BONDSTER_BCO_#{k}=#{v}"}]
  config = config.join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{config}' > /etc/init/bondster-bco.conf)

  ids = %x(systemctl list-units | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("bondster-bco-")
  }.map { |x| x.chomp(".service") }

  ids.each { |e|
    %x(systemctl restart #{e} 2>&1)
  }

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end
