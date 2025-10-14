local setup_crafter = require 'setup_crafter'
local setup_storage = require 'setup_storage'
local setup_v2 = require 'setup_v2'

local network = 'net1'

-- local custom_craft_storage = testnet.createInventory(network, 'custom_storage')
-- local custom_craft_tank = testnet.createTank(network, {
--     name = 'custom_tank',
--     tanks = { 1000, 1000 },
-- })
-- testnet.setFluid(custom_craft_tank, {
--     name = 'minecraft:lava',
--     amount = 1000,
-- })

-- setup_storage(network, 1)
-- setup_crafter(network, 2)
-- setup_v2(network, 3)

-- local m1 = testnet.createModem(network)
ccemux.openEmu(4)
-- testnet.attachPeripheral('top', m1, 3)

ccemux.closeEmu()
