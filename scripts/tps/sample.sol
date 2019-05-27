pragma solidity ^0.4.0;

contract Sample {
    int32 balance;
   
    function add() {
        balance += 1;
    }
    
    function get(string _key) returns(int32){
        return balance;
    }
}
