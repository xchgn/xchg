module xchg::xchg;

use std::string::String;
use sui::table::{Self, Table};

public struct RouterInfo has key, store {
    id: UID,
    name: String,
    XchgAddress: address,
    NativeAddress: String,
}

public struct ClientInfo has key, store {
    id: UID,
    name: String,
    XchgAddress: address,
}

public struct Routers has store {
    routers: Table<address, RouterInfo>
}

public struct Clients has store {
    clients: Table<address, ClientInfo>
}

public struct Storage has key {
    id: UID,
    routers: Routers,
    clients: Clients,
}

fun init(ctx: &mut TxContext) {
    let clients_table = table::new<address, ClientInfo>(ctx);
    let routers_table = table::new<address, RouterInfo>(ctx);
    let clients = Clients { clients: clients_table };
    let routers = Routers { routers: routers_table };
    transfer::share_object(Storage {
			id: object::new(ctx),
			clients: clients,
            routers: routers,
		})
}

public fun declare_router(storage : &mut Storage, name: String, xchgAddress: address, nativeAddress: String, ctx: &mut TxContext) {
    let router = RouterInfo {
        id: object::new(ctx),
        name: name,
        XchgAddress: xchgAddress,
        NativeAddress: nativeAddress,
    };
    storage.routers.routers.add(xchgAddress, router);
}

public fun declare_client(storage : &mut Storage, name: String, xchgAddress: address, ctx: &mut TxContext) {
    let client = ClientInfo {
        id: object::new(ctx),
        name: name,
        XchgAddress: xchgAddress,
    };
    storage.clients.clients.add(xchgAddress, client);
}
