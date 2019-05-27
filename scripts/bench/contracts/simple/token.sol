pragma solidity ^0.4.21;

contract Token {
    mapping (address => uint) public balances;

    function Token() public{
        balances[msg.sender] = 1000000;
    }

    function transfer(address _to, uint _amount) public{
        if (balances[msg.sender] < _amount) {
            return;
        }

        balances[msg.sender] -= _amount;
        balances[_to] += _amount;
    }
}