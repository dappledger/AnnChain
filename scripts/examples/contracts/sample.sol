pragma solidity ^0.4.0;

contract Civilwar {
    int public simpleValue;
    address public owner;
    
    function Civilwar() {
        owner = msg.sender;
        simpleValue = -123;
    }
    
    function setOwner(address newOwner) returns (address old) {
        old = owner;
        owner = newOwner;
    }
    
    function setSimpleValue(int v) {
        simpleValue = v;
    }
    
    function playASM() {
        assembly{
            
        }
    }
}
