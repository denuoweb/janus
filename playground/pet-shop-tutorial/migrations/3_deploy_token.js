var HRC20 = artifacts.require("HRC20Token");

module.exports = async function(deployer) {
  await deployer.deploy(HRC20);
};
