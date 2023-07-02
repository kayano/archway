# ADR-007 Contract Registry Module

## Status

Draft

## Abstract

We propose a new module called `x/contractregistry` which can be used by contract deployers to provide metadata about their contracts. 
The module would store contract metadata such as source code, schema, contact info.

## Context

Cosmwasm does not provide any way for a Cosmwasm smart contract developer to provide any metadata regarding their contracts. This has been explored in `x/wasm` before [^1], where during contract upload, a developer could provide source code url. This feature was deprecated by Confio due to
1. Field was often unfilled or had incorrect values
2. No tooling to verify the contracts match the given information   
Due to the nature of wasm, it is also not possible to take a look at the source code of a deployed contract.

Once a contract is deployed, it is not easy to get the contract schema for external parties to get the contract endpoints[^2], especially so in the case when the source of the contract is closed source or source URL not available. Having this information available on chain would enable the following
1. Make third party tooling of contracts easier to develop, such as code gen for UI
2. General purpose contract interaction tools
3. Indexers and block exploreres could use this information to better store and display the contract state.  

Currently, there is no way for a user/another developer to know who deployed a contract and in case they would like to contact the developer, there isnt any way to do it beyond the deployer address. Adding a field for security contact would help other report issues.

Most of the Cosmwasm chains run as permissioned Cosmwasm, which allows for the contract source to be connected to the binary in the governance proposal. However, in the permissionless approach of Archway, there is no builtin way to establish this connection.

## Architecture

The solution proposed is the develop a new sdk module `x/contractregistry` which will store the relevant information.
The feature will be an opt in where developers can choose to provide only the necessary info that they deem important. e.g A developer might want to share the schema to allow external tools but might not want to share their source code url as their code is closed source.

### Why module instead of modifying wasmd?

1. To keep the archway-wasmd fork as simple in diff as possible so that its easier to upgrade to later versions of wasmd.
2. `x/wasmd` is purpose scoped to be a contract execution engine. Expanding the module to also include code registry features will make the module inflexible.

### Why module instead of a name service contract?

The features required for the contract registry could be built either as a chain module or a smart contract. However, going by the philosophy of what a smart contract is, which is developed to run dapps, it does not make sense to put this feature in a smart contract as this is meant to be a metadata service which augments on the existing smart contract functionality provided by the chain. This feature is tighly coupled with the on chain contract management to be deployed as a smart contract.

### Technical Specification

The module would store the following state for Code which has been deployed on chain:
```proto
message CodeMetadata {
    // The Code ID of the deployed contract
    uint64 code_id = 1;
    // The information regarding the contract source codebase
    SourceMetadata source = 2;
    // The information regarding the image used to build and optimize the contract binary
    SourceBuilder source_builder = 3;
    // The JSON schema which specifies the interaction endpoints of the contract
    string schema = 4;
    // The contacts of the developers or security incidence handlers 
    repeated string contacts = 5;
}

message SourceMetadata {
    // The link to the code repository. e.g https://github.com/archway-network/archway
    string repository = 1;
    // The tag of the commit message at which the binary was built and deployed. e.g v1.0.2
    string tag = 2;
    // The software license of the smart contract code. e.g Apache-2.0
    string license = 3;
}

message SourceBuilder {
    // Docker image. e.g cosmwasm/rust-optimizer
    string image = 1; 
    // Docker image tag. e.g 0.12.6
    string tag = 2;
    // Name of the generated contract binary. e.g counter.wasm
    string contract_name = 3;
}
```
This information can only be modified by the user who uploaded the contract binary. Even though Code IDs are unique to binary, we should make these fields modifyable over time to allow for fixing erroneous values and updating contacts.

## Consequences 

### Backwards compatibility

Since the feature is being added as a new module, this should not cause any backwards compatibility issues.

### Positive

1. Sets the groundwork for more comprehensive developer tooling
2. Contracts that have registered with source code tag might be given more trusted access, such as access to Begin and Endblockers.

### Negative

1. The total storage of the chain would increase, especially when storing the contract schema.
2. The image builder is not executed to test against the source code. This would mean the responsiblity to test if the provided information matches will be provided by off-chain tooling. 

### Neutral

1. Tools like archwaycli would have to be modified to support this feature.
 


[^1]: [Question: Why was StoreCode.url removed from the tx msg?](https://github.com/CosmWasm/wasmd/issues/742)

[^2]: [Upload JSON schema alongside code contract](https://github.com/CosmWasm/wasmd/issues/241)