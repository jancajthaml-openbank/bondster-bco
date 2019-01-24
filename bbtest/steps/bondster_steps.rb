
step "bondster gateway contains following statements" do |statements|
  statements = JSON.parse(statements)

  BondsterMock.set_statements(statements)
end
