/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import chai from "chai";
import chaiAsPromised from "chai-as-promised";
import { HTTPSigner } from "../../../src";
import MockDate from "mockdate";

chai.use(chaiAsPromised);
const expect = chai.expect;

describe("HTTP Signature Client - Constructor", function () {
  it("throws an error constructing signer instance with empty signingKey", () => {
    expect(() => new HTTPSigner({ signingKey: null })).to.throw(
      TypeError,
      "Error initializing HTTPSigner: signingKey cannot be empty"
    );
  });
  it("throws an error constructing signer instance when privateKey is missing", () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      kid: "mock-kid",
    };

    expect(() => new HTTPSigner({ signingKey: mockSigningKey })).to.throw(
      TypeError,
      "Error initializing HTTPSigner: privateKey cannot be empty"
    );
  });
  it("throws an error constructing signer instance when publicKey is missing", () => {
    const mockSigningKey = {
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };

    expect(() => new HTTPSigner({ signingKey: mockSigningKey })).to.throw(
      TypeError,
      "Error initializing HTTPSigner: publicKey cannot be empty"
    );
  });
  it("throws an error constructing signer instance when kid is missing", () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
    };

    expect(() => new HTTPSigner({ signingKey: mockSigningKey })).to.throw(
      TypeError,
      "Error initializing HTTPSigner: kid cannot be empty"
    );
  });
  it("successfully constructs signer instance", () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    expect(signer).to.be.an.instanceOf(HTTPSigner);
    expect(signer).to.have.property("signingKey");
    expect(signer.signingKey).to.have.property("publicKey");
    expect(signer.signingKey.publicKey).to.equal("mock-public-key");
    expect(signer.signingKey).to.have.property("privateKey");
    expect(signer.signingKey.privateKey).to.equal("mock-private-key");
    expect(signer.signingKey).to.have.property("kid");
    expect(signer.signingKey.kid).to.equal("mock-kid");
  });
  it("successfully constructs signer instance with authorization", () => {
    const mockAuthorization = "mock-authorization";
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const signer = new HTTPSigner({
      authorization: mockAuthorization,
      signingKey: mockSigningKey,
    });

    expect(signer).to.be.an.instanceOf(HTTPSigner);
    expect(signer).to.have.property("authorization");
    expect(signer.authorization).to.equal("mock-authorization");
    expect(signer).to.have.property("signingKey");
    expect(signer.signingKey).to.have.property("publicKey");
    expect(signer.signingKey.publicKey).to.equal("mock-public-key");
    expect(signer.signingKey).to.have.property("privateKey");
    expect(signer.signingKey.privateKey).to.equal("mock-private-key");
    expect(signer.signingKey).to.have.property("kid");
    expect(signer.signingKey.kid).to.equal("mock-kid");
  });
});
describe("HTTP Signature Client - Generating Parameters", function () {
  MockDate.set("1999-04-03");
  it("successfully generates signature params", () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockCreated = Math.floor(Date.now() / 1000);
    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    const sigParams = signer.generateSignatureParams();

    expect(sigParams).to.be.a("string");
    expect(sigParams).to.equal(
      `("@method" "@target-uri" "content-digest");created=${mockCreated};keyid="${signer.signingKey.kid}"`
    );
  });
  it("successfully generates signature params with authorization", () => {
    const mockAuthorization = "mock-authorization";
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockCreated = Math.floor(Date.now() / 1000);
    const signer = new HTTPSigner({
      authorization: mockAuthorization,
      signingKey: mockSigningKey,
    });

    const sigParams = signer.generateSignatureParams();

    expect(sigParams).to.be.a("string");
    expect(sigParams).to.equal(
      `("@method" "@target-uri" "content-digest" "authorization");created=${mockCreated};keyid="${signer.signingKey.kid}"`
    );
  });
});
describe("HTTP Signature Client - Getting Signature Input", function () {
  it("successfully gets signature input", () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockSigName = "mock-sig-name";
    const mockSigParams = `("@method" "@target-uri" "content-digest" "authorization");created=123456789;keyid="mock-kid"`;

    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    const sigInput = signer.getSignatureInput(mockSigName, mockSigParams);

    expect(sigInput).to.be.a("string");
    expect(sigInput).to.equal(`${mockSigName}=${mockSigParams}`);
  });
  it("throws an error when name is missing", () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockSigParams = `("@method" "@target-uri" "content-digest" "authorization");created=123456789;keyid="mock-kid"`;

    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    expect(() => signer.getSignatureInput(null, mockSigParams)).to.throw(
      TypeError,
      "Error getting signature input: name is required"
    );
  });
  it("throws an error when signature params are missing", () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockName = "mock-name";

    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    expect(() => signer.getSignatureInput(mockName, null)).to.throw(
      TypeError,
      "Error getting signature input: signature parameters are required"
    );
  });
});
describe("HTTP Signature Client - Signing", function () {
  it("throws an error when digest is missing", async () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockUrl = "mock-url";
    const mockSigName = "mock-sig-name";
    const mockSigParams = "mock-sig-params";

    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    expect(
      signer.sign({
        digest: null,
        url: mockUrl,
        name: mockSigName,
        sigParams: mockSigParams,
      })
    ).to.eventually.throw(
      TypeError,
      "Error generating a signature: digest is missing"
    );
  });
  it("throws an error when url is missing", async () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockDigest = "mock-digest";
    const mockSigName = "mock-sig-name";
    const mockSigParams = "mock-sig-params";

    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    expect(
      signer.sign(mockDigest, null, mockSigName, mockSigParams)
    ).to.eventually.throw(
      TypeError,
      "Error generating a signature: server url is missing"
    );
  });
  it("throws an error when signature name is missing", async () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockDigest = "mock-digest";
    const mockUrl = "mock-url";
    const mockSigParams = "mock-sig-params";

    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    expect(
      signer.sign(mockDigest, mockUrl, null, mockSigParams)
    ).to.eventually.throw(
      TypeError,
      "Error generating a signature: name is missing"
    );
  });
  it("throws an error when signature parameters are missing", async () => {
    const mockSigningKey = {
      publicKey: "mock-public-key",
      privateKey: "mock-private-key",
      kid: "mock-kid",
    };
    const mockDigest = "mock-digest";
    const mockUrl = "mock-url";
    const mockSigName = "mock-sig-name";

    const signer = new HTTPSigner({ signingKey: mockSigningKey });

    expect(
      signer.sign(mockDigest, mockUrl, mockSigName, null)
    ).to.eventually.throw(
      TypeError,
      "Error generating a signature: signature parameters are missing"
    );
  });
});
