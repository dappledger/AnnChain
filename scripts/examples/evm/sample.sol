pragma solidity ^0.4.0;

contract Simple {
    mapping(string=>string) kvData;
   
    function set(string _key, string _value) {
        kvData[_key] = _value;
    }
    
    function get(string _key) returns(string){
        return kvData[_key];
    }
}
