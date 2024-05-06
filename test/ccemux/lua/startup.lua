local m1 = testnet.createModem(1)
testnet.attachPeripheral('top', m1)

local i1 = testnet.createInventory(1)
local i2 = testnet.createInventory(1)

testnet.setItem(i1, 1, {
    name = 'minecraft:apple',
    count = 2,
    maxCount = 64,
})

testnet.setItem(i2, 2, {
    name = 'minecraft:ender_pearl',
    count = 3,
    maxCount = 16,
})

require 'startup2'
