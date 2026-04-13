import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";

interface Props {
  initialSettings: Record<string, unknown>;
  secretsSet?: Record<string, boolean>;
  onSave: (settings: Record<string, unknown>) => Promise<void>;
  onCancel: () => void;
}

export function WebSearchChainForm({ initialSettings, secretsSet, onSave, onCancel }: Props) {
  const { t } = useTranslation("tools");
  const tavily = useMemo(
    () => ((initialSettings.tavily ?? {}) as Record<string, unknown>),
    [initialSettings],
  );
  const [enabled, setEnabled] = useState(Boolean(tavily.enabled ?? true));
  const [maxResults, setMaxResults] = useState(Number(tavily.max_results ?? 5));
  const [apiKey, setApiKey] = useState("");
  const [showKeyInput, setShowKeyInput] = useState(false);
  const [saving, setSaving] = useState(false);

  const secretKey = "tools.web.tavily.api_key";
  const keyIsSet = secretsSet?.[secretKey] === true;
  const showInput = showKeyInput || !keyIsSet;

  const handleSave = async () => {
    setSaving(true);
    try {
      const settings: Record<string, unknown> = {
        tavily: {
          enabled,
          max_results: maxResults,
          ...(apiKey.trim() !== "" ? { api_key: apiKey.trim() } : {}),
        },
      };
      await onSave(settings);
    } catch {
      // toast shown by hook
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <DialogHeader>
        <DialogTitle>{t("builtin.searchChain.title")}</DialogTitle>
        <DialogDescription>{t("builtin.searchChain.description")}</DialogDescription>
      </DialogHeader>

      <div className="my-4 space-y-4">
        <div className="rounded-lg border bg-card p-4 space-y-4">
          <div className="flex items-center gap-3">
            <Switch size="sm" checked={enabled} onCheckedChange={setEnabled} />
            <div>
              <div className="text-sm font-medium">{t("builtin.searchChain.providers.tavily")}</div>
              <div className="text-xs text-muted-foreground">{t("builtin.searchChain.providerHint")}</div>
            </div>
          </div>

          <div className="grid gap-1.5 max-w-32">
            <Label className="text-xs text-muted-foreground">{t("builtin.searchChain.maxResults")}</Label>
            <Input
              type="number"
              min={1}
              max={10}
              value={maxResults}
              onChange={(e) => setMaxResults(Number(e.target.value))}
              className="text-base md:text-sm"
            />
          </div>

          <div className="grid gap-1.5">
            <Label className="text-xs text-muted-foreground">{t("builtin.searchChain.apiKey")}</Label>
            {keyIsSet && !showInput ? (
              <div className="flex items-center gap-2">
                <span className="text-xs font-medium text-green-600 dark:text-green-400">
                  {t("builtin.searchChain.apiKeySet")}
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 px-2 text-xs"
                  onClick={() => setShowKeyInput(true)}
                >
                  {t("builtin.searchChain.apiKeyChange")}
                </Button>
              </div>
            ) : (
              <Input
                type="password"
                autoComplete="off"
                placeholder={
                  keyIsSet
                    ? t("builtin.searchChain.apiKeyReplacePlaceholder")
                    : t("builtin.searchChain.apiKeyPlaceholder")
                }
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                className="text-base md:text-sm font-mono"
              />
            )}
            {showInput && <p className="text-xs text-muted-foreground">{t("builtin.searchChain.apiKeyHint")}</p>}
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" onClick={onCancel}>
          {t("builtin.searchChain.cancel")}
        </Button>
        <Button onClick={handleSave} disabled={saving}>
          {saving && <Loader2 className="h-4 w-4 animate-spin" />}
          {saving ? t("builtin.searchChain.saving") : t("builtin.searchChain.save")}
        </Button>
      </DialogFooter>
    </div>
  );
}
