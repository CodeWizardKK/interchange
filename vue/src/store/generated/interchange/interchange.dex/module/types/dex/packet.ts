/* eslint-disable */
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "interchange.dex";

export interface DexPacketData {
  noData: NoData | undefined;
  /** this line is used by starport scaffolding # ibc/packet/proto/field */
  createPairPacket: CreatePairPacketData | undefined;
}

export interface NoData {}

/** CreatePairPacketData defines a struct for the packet payload */
export interface CreatePairPacketData {
  sourceDenom: string;
  targetDenom: string;
}

/** CreatePairPacketAck defines a struct for the packet acknowledgment */
export interface CreatePairPacketAck {}

const baseDexPacketData: object = {};

export const DexPacketData = {
  encode(message: DexPacketData, writer: Writer = Writer.create()): Writer {
    if (message.noData !== undefined) {
      NoData.encode(message.noData, writer.uint32(10).fork()).ldelim();
    }
    if (message.createPairPacket !== undefined) {
      CreatePairPacketData.encode(
        message.createPairPacket,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): DexPacketData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseDexPacketData } as DexPacketData;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.noData = NoData.decode(reader, reader.uint32());
          break;
        case 2:
          message.createPairPacket = CreatePairPacketData.decode(
            reader,
            reader.uint32()
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DexPacketData {
    const message = { ...baseDexPacketData } as DexPacketData;
    if (object.noData !== undefined && object.noData !== null) {
      message.noData = NoData.fromJSON(object.noData);
    } else {
      message.noData = undefined;
    }
    if (
      object.createPairPacket !== undefined &&
      object.createPairPacket !== null
    ) {
      message.createPairPacket = CreatePairPacketData.fromJSON(
        object.createPairPacket
      );
    } else {
      message.createPairPacket = undefined;
    }
    return message;
  },

  toJSON(message: DexPacketData): unknown {
    const obj: any = {};
    message.noData !== undefined &&
      (obj.noData = message.noData ? NoData.toJSON(message.noData) : undefined);
    message.createPairPacket !== undefined &&
      (obj.createPairPacket = message.createPairPacket
        ? CreatePairPacketData.toJSON(message.createPairPacket)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<DexPacketData>): DexPacketData {
    const message = { ...baseDexPacketData } as DexPacketData;
    if (object.noData !== undefined && object.noData !== null) {
      message.noData = NoData.fromPartial(object.noData);
    } else {
      message.noData = undefined;
    }
    if (
      object.createPairPacket !== undefined &&
      object.createPairPacket !== null
    ) {
      message.createPairPacket = CreatePairPacketData.fromPartial(
        object.createPairPacket
      );
    } else {
      message.createPairPacket = undefined;
    }
    return message;
  },
};

const baseNoData: object = {};

export const NoData = {
  encode(_: NoData, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): NoData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseNoData } as NoData;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): NoData {
    const message = { ...baseNoData } as NoData;
    return message;
  },

  toJSON(_: NoData): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<NoData>): NoData {
    const message = { ...baseNoData } as NoData;
    return message;
  },
};

const baseCreatePairPacketData: object = { sourceDenom: "", targetDenom: "" };

export const CreatePairPacketData = {
  encode(
    message: CreatePairPacketData,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.sourceDenom !== "") {
      writer.uint32(10).string(message.sourceDenom);
    }
    if (message.targetDenom !== "") {
      writer.uint32(18).string(message.targetDenom);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): CreatePairPacketData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseCreatePairPacketData } as CreatePairPacketData;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sourceDenom = reader.string();
          break;
        case 2:
          message.targetDenom = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CreatePairPacketData {
    const message = { ...baseCreatePairPacketData } as CreatePairPacketData;
    if (object.sourceDenom !== undefined && object.sourceDenom !== null) {
      message.sourceDenom = String(object.sourceDenom);
    } else {
      message.sourceDenom = "";
    }
    if (object.targetDenom !== undefined && object.targetDenom !== null) {
      message.targetDenom = String(object.targetDenom);
    } else {
      message.targetDenom = "";
    }
    return message;
  },

  toJSON(message: CreatePairPacketData): unknown {
    const obj: any = {};
    message.sourceDenom !== undefined &&
      (obj.sourceDenom = message.sourceDenom);
    message.targetDenom !== undefined &&
      (obj.targetDenom = message.targetDenom);
    return obj;
  },

  fromPartial(object: DeepPartial<CreatePairPacketData>): CreatePairPacketData {
    const message = { ...baseCreatePairPacketData } as CreatePairPacketData;
    if (object.sourceDenom !== undefined && object.sourceDenom !== null) {
      message.sourceDenom = object.sourceDenom;
    } else {
      message.sourceDenom = "";
    }
    if (object.targetDenom !== undefined && object.targetDenom !== null) {
      message.targetDenom = object.targetDenom;
    } else {
      message.targetDenom = "";
    }
    return message;
  },
};

const baseCreatePairPacketAck: object = {};

export const CreatePairPacketAck = {
  encode(_: CreatePairPacketAck, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): CreatePairPacketAck {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseCreatePairPacketAck } as CreatePairPacketAck;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): CreatePairPacketAck {
    const message = { ...baseCreatePairPacketAck } as CreatePairPacketAck;
    return message;
  },

  toJSON(_: CreatePairPacketAck): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<CreatePairPacketAck>): CreatePairPacketAck {
    const message = { ...baseCreatePairPacketAck } as CreatePairPacketAck;
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;
