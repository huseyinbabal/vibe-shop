import { Text, TextInput, View, type TextInputProps } from "react-native"

// Field is the shared zinc-styled labelled input used by the auth forms.
export function Field({ label, ...props }: TextInputProps & { label: string }) {
  return (
    <View className="gap-2">
      <Text className="text-sm font-medium text-zinc-900">{label}</Text>
      <TextInput
        className="h-12 rounded-md border border-zinc-200 bg-white px-3 text-base text-zinc-900"
        placeholderTextColor="#a1a1aa"
        autoCapitalize="none"
        {...props}
      />
    </View>
  )
}
