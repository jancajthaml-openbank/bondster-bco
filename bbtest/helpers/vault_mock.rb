module VaultMock

  class << self
    attr_accessor :tenants
  end

  self.tenants = Hash.new()

  def self.reset()
    self.tenants = Hash.new()
  end

  def self.get_tenants()
    self.tenants.keys
  end

  def self.get_accounts(tenant)
    return nil unless self.tenants.has_key?(tenant)
    return self.tenants[tenant].keys
  end

  def self.get_acount(tenant, id)
    return {} unless self.tenants.has_key?(tenant)
    return {} unless self.tenants[tenant].has_key?(id)
    return self.tenants[tenant][id]
  end

  def self.create_account(tenant, id, format, currency, is_balance_check)
    return if self.tenants.has_key?(tenant) && self.tenants[tenant].has_key?(id)
    self.tenants[tenant] = Hash.new() unless self.tenants.has_key?(tenant)
    self.tenants[tenant][id] = {
      :currency => currency,
      :format => format,
      :isBalanceCheck => is_balance_check,
    }
    return
  end

end
