import { useFeatureSupport } from "@canva/app-hooks";
import {
  Alert,
  Button,
  FormField,
  LoadingIndicator,
  Rows,
  Text,
  TextInput,
  Title,
} from "@canva/app-ui-kit";
import { upload } from "@canva/asset";
import { addElementAtCursor, addElementAtPoint } from "@canva/design";
import { requestOpenExternalUrl } from "@canva/platform";
import { auth } from "@canva/user";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useIntl } from "react-intl";
import * as styles from "styles/components.css";
import type { ShortLink } from "../../trunc";
import {
  isWebAddress,
  qrDataUrl,
  shortenUrl,
  TRUNC_SCOPE,
  truncBaseUrl,
  TruncError,
} from "../../trunc";

type Status = "checking" | "signed-out" | "signed-in";

// A URL scheme, not text to translate.
const URL_PLACEHOLDER = "https://";

export const App = () => {
  const intl = useIntl();
  const oauth = useMemo(() => auth.initOauth(), []);
  const isSupported = useFeatureSupport();
  const addElement = [addElementAtPoint, addElementAtCursor].find((fn) =>
    isSupported(fn),
  );

  const [status, setStatus] = useState<Status>("checking");
  const [url, setUrl] = useState("");
  const [link, setLink] = useState<ShortLink | null>(null);
  const [busy, setBusy] = useState<"shorten" | "qr" | "text" | "signin" | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);

  /** The current token; Canva caches and refreshes it. */
  const token = useCallback(async () => {
    const res = await oauth.getAccessToken({ scope: TRUNC_SCOPE });
    return res?.token ?? null;
  }, [oauth]);

  useEffect(() => {
    token()
      .then((t) => setStatus(t ? "signed-in" : "signed-out"))
      .catch(() => setStatus("signed-out"));
  }, [token]);

  const signIn = async () => {
    setBusy("signin");
    setError(null);
    try {
      await oauth.requestAuthorization({ scope: TRUNC_SCOPE });
      setStatus((await token()) ? "signed-in" : "signed-out");
    } catch {
      setError(
        intl.formatMessage({
          defaultMessage: "Signing in did not finish. Please try again.",
          description: "Error shown when Sign in with Trunc fails or is cancelled.",
        }),
      );
    } finally {
      setBusy(null);
    }
  };

  const signOut = async () => {
    await oauth.deauthorize();
    setLink(null);
    setStatus("signed-out");
  };

  const shorten = async () => {
    setError(null);
    if (!isWebAddress(url)) {
      setError(
        intl.formatMessage({
          defaultMessage: "Enter a full web address, starting with https://.",
          description: "Validation error for the web address field.",
        }),
      );
      return;
    }
    setBusy("shorten");
    try {
      const t = await token();
      if (!t) {
        setStatus("signed-out");
        return;
      }
      setLink(await shortenUrl(t, url));
    } catch (e) {
      if (e instanceof TruncError && e.status === 401) {
        setStatus("signed-out");
        return;
      }
      setError(
        e instanceof TruncError && e.status !== 0
          ? e.message
          : intl.formatMessage({
              defaultMessage: "Couldn't reach Trunc. Check your connection.",
              description: "Error shown when the Trunc server can't be reached.",
            }),
      );
    } finally {
      setBusy(null);
    }
  };

  const addQr = async () => {
    if (!link || !addElement) return;
    setBusy("qr");
    try {
      const dataUrl = await qrDataUrl(link.shortUrl);
      const { ref } = await upload({
        type: "image",
        mimeType: "image/png",
        url: dataUrl,
        thumbnailUrl: dataUrl,
        aiDisclosure: "none",
      });
      await addElement({
        type: "image",
        ref,
        altText: {
          text: intl.formatMessage(
            {
              defaultMessage: "QR code for {url}",
              description: "Alt text of the QR code image added to the design.",
            },
            { url: link.shortUrl },
          ),
          decorative: false,
        },
      });
    } finally {
      setBusy(null);
    }
  };

  const addText = async () => {
    if (!link || !addElement) return;
    setBusy("text");
    try {
      await addElement({ type: "text", children: [link.shortUrl] });
    } finally {
      setBusy(null);
    }
  };

  if (status === "checking") {
    return (
      <div className={styles.scrollContainer}>
        <LoadingIndicator />
      </div>
    );
  }

  if (status === "signed-out") {
    return (
      <div className={styles.scrollContainer}>
        <Rows spacing="2u">
          <Title size="small">
            {intl.formatMessage({
              defaultMessage: "Short links and QR codes",
              description: "Heading of the signed-out panel.",
            })}
          </Title>
          <Text>
            {intl.formatMessage({
              defaultMessage:
                "Make a short link, add its QR code to your design, and see how many people scan it in Trunc.",
              description: "What the app does, shown before signing in.",
            })}
          </Text>
          {error && <Alert tone="critical">{error}</Alert>}
          <Button
            variant="primary"
            onClick={signIn}
            loading={busy === "signin"}
            stretch
          >
            {intl.formatMessage({
              defaultMessage: "Sign in with Trunc",
              description: "Button that starts Sign in with Trunc.",
            })}
          </Button>
        </Rows>
      </div>
    );
  }

  const unsupported = intl.formatMessage({
    defaultMessage: "This can't be added to the current page",
    description: "Tooltip when elements can't be added to the current design.",
  });

  return (
    <div className={styles.scrollContainer}>
      <Rows spacing="2u">
        <FormField
          label={intl.formatMessage({
            defaultMessage: "Web address",
            description: "Label of the field for the link to shorten.",
          })}
          value={url}
          control={(props) => (
            <TextInput
              {...props}
              placeholder={URL_PLACEHOLDER}
              onChange={(v) => setUrl(v)}
            />
          )}
        />
        {error && <Alert tone="critical">{error}</Alert>}
        <Button
          variant="primary"
          onClick={shorten}
          loading={busy === "shorten"}
          stretch
        >
          {intl.formatMessage({
            defaultMessage: "Create short link",
            description: "Button that shortens the web address.",
          })}
        </Button>

        {link && (
          <Rows spacing="1u">
            <Text>
              {intl.formatMessage({
                defaultMessage: "Your short link",
                description: "Label above the new short link.",
              })}
            </Text>
            <Text variant="bold">{link.shortUrl}</Text>
            <Button
              variant="secondary"
              onClick={addQr}
              loading={busy === "qr"}
              disabled={!addElement}
              tooltipLabel={!addElement ? unsupported : undefined}
              stretch
            >
              {intl.formatMessage({
                defaultMessage: "Add QR code to design",
                description: "Button that adds the link's QR code as an image.",
              })}
            </Button>
            <Button
              variant="secondary"
              onClick={addText}
              loading={busy === "text"}
              disabled={!addElement}
              tooltipLabel={!addElement ? unsupported : undefined}
              stretch
            >
              {intl.formatMessage({
                defaultMessage: "Add link as text",
                description: "Button that adds the short link as a text element.",
              })}
            </Button>
          </Rows>
        )}

        <Button
          variant="tertiary"
          onClick={() =>
            requestOpenExternalUrl({ url: `${truncBaseUrl()}/dashboard` })
          }
        >
          {intl.formatMessage({
            defaultMessage: "See clicks in Trunc",
            description: "Opens the Trunc dashboard.",
          })}
        </Button>
        <Button variant="tertiary" onClick={signOut}>
          {intl.formatMessage({
            defaultMessage: "Sign out",
            description: "Signs out of Trunc.",
          })}
        </Button>
      </Rows>
    </div>
  );
};
